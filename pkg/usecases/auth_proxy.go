package usecases

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/network"
	"github.com/int128/kauthproxy/pkg/portforwarder"
	"github.com/int128/kauthproxy/pkg/resolver"
	"github.com/int128/kauthproxy/pkg/reverseproxy"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

var Set = wire.NewSet(
	wire.Struct(new(AuthProxy), "*"),
	wire.Bind(new(AuthProxyInterface), new(*AuthProxy)),
)

type AuthProxyInterface interface {
	Do(ctx context.Context, in AuthProxyOptions) error
}

// AuthProxy provides a use-case of authentication proxy.
type AuthProxy struct {
	ReverseProxy    reverseproxy.Interface
	PortForwarder   portforwarder.Interface
	ResolverFactory resolver.FactoryInterface
	Network         network.Interface
}

// AuthProxyOptions represents an option of AuthProxy.
type AuthProxyOptions struct {
	Config    *rest.Config
	Namespace string
	RemoteURL *url.URL
	LocalAddr string
}

// Do runs the use-case.
func (u *AuthProxy) Do(ctx context.Context, o AuthProxyOptions) error {
	rsv, err := u.ResolverFactory.New(o.Config)
	if err != nil {
		return xerrors.Errorf("could not create a resolver: %w", err)
	}
	pod, containerPort, err := parseRemoteURL(rsv, o.Namespace, o.RemoteURL)
	if err != nil {
		return xerrors.Errorf("could not find the pod and container port: %w", err)
	}
	transitPort, err := u.Network.AllocateLocalPort()
	if err != nil {
		return xerrors.Errorf("could not allocate a local port: %w", err)
	}
	transport, err := u.Network.NewTransportWithToken(o.Config)
	if err != nil {
		return xerrors.Errorf("could not create a transport for reverse proxy: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	u.ReverseProxy.Start(ctx, eg,
		reverseproxy.Options{
			Transport: transport,
			Source:    reverseproxy.Source{Address: o.LocalAddr},
			Target: reverseproxy.Target{
				Scheme: o.RemoteURL.Scheme,
				Host:   "localhost",
				Port:   transitPort,
			},
		})
	if err := u.PortForwarder.Start(ctx, eg,
		portforwarder.Options{
			Config: o.Config,
			Source: portforwarder.Source{Port: transitPort},
			Target: portforwarder.Target{
				Pod:           pod,
				ContainerPort: containerPort,
			},
		}); err != nil {
		return xerrors.Errorf("could not start a port forwarder: %w", err)
	}
	log.Printf("Open http://%s", o.LocalAddr)
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while port-forwarding: %w", err)
	}
	return nil
}

func parseRemoteURL(r resolver.Interface, namespace string, u *url.URL) (*v1.Pod, int, error) {
	h := u.Hostname()
	if strings.HasSuffix(h, ".svc") {
		serviceName := strings.TrimSuffix(h, ".svc")
		return r.FindByServiceName(namespace, serviceName)
	}
	return r.FindByPodName(namespace, h)
}
