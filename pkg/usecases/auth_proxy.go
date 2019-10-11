package usecases

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/logger"
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
	Logger          logger.Interface
}

// AuthProxyOptions represents an option of AuthProxy.
type AuthProxyOptions struct {
	Config                *rest.Config
	Namespace             string
	TargetURL             *url.URL
	BindAddressCandidates []string
}

// Do runs the use-case.
func (u *AuthProxy) Do(ctx context.Context, o AuthProxyOptions) error {
	rsv, err := u.ResolverFactory.New(o.Config)
	if err != nil {
		return xerrors.Errorf("could not create a resolver: %w", err)
	}
	pod, containerPort, err := parseTargetURL(rsv, o.Namespace, o.TargetURL)
	if err != nil {
		return xerrors.Errorf("could not find the pod and container port: %w", err)
	}
	u.Logger.V(1).Infof("found container port %d of pod %s", containerPort, pod.Name)
	transitPort, err := u.Network.AllocateLocalPort()
	if err != nil {
		return xerrors.Errorf("could not allocate a local port: %w", err)
	}
	u.Logger.V(1).Infof("client -> reverse_proxy -> port_forwarder:%d -> pod -> container:%d", transitPort, containerPort)
	transport, err := u.Network.NewTransportWithToken(o.Config)
	if err != nil {
		return xerrors.Errorf("could not create a transport for reverse proxy: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		rpo := reverseproxy.Options{
			Transport: transport,
			Source: reverseproxy.Source{
				AddressCandidates: o.BindAddressCandidates,
			},
			Target: reverseproxy.Target{
				Scheme: o.TargetURL.Scheme,
				Host:   "localhost",
				Port:   transitPort,
			},
		}
		if err := u.ReverseProxy.Run(ctx, rpo); err != nil {
			return xerrors.Errorf("could not run a reverse proxy: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		for {
			pfo := portforwarder.Options{
				Config: o.Config,
				Source: portforwarder.Source{
					Port: transitPort,
				},
				Target: portforwarder.Target{
					Pod:           pod,
					ContainerPort: containerPort,
				},
			}
			if err := u.PortForwarder.Run(ctx, pfo); err != nil {
				return xerrors.Errorf("could not run a port forwarder: %w", err)
			}
			// retry connection
		}
	})
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while running an authentication proxy: %w", err)
	}
	return nil
}

func parseTargetURL(r resolver.Interface, namespace string, u *url.URL) (*v1.Pod, int, error) {
	h := u.Hostname()
	if strings.HasSuffix(h, ".svc") {
		serviceName := strings.TrimSuffix(h, ".svc")
		return r.FindByServiceName(namespace, serviceName)
	}
	return r.FindByPodName(namespace, h)
}
