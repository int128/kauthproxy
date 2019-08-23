package usecases

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/portforwarder"
	"github.com/int128/kauthproxy/pkg/reverseproxy"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/plugin/pkg/client/auth/exec"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
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
	ReverseProxy  reverseproxy.Interface
	PortForwarder portforwarder.Interface
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
	r, err := newResolver(o.Config)
	if err != nil {
		return xerrors.Errorf("could not create a resolver: %w", err)
	}
	pod, containerPort, err := parseRemoteURL(r, o.Namespace, o.RemoteURL)
	if err != nil {
		return xerrors.Errorf("could not find the pod and container port: %w", err)
	}

	transitPort, err := findFreePort()
	if err != nil {
		return xerrors.Errorf("could not allocate a local port: %w", err)
	}
	authProxyTransport, err := newAuthProxyTransport(o.Config)
	if err != nil {
		return xerrors.Errorf("could not create a transport for reverse proxy: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	u.ReverseProxy.Start(ctx, eg,
		reverseproxy.Options{
			Transport: authProxyTransport,
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

func newAuthProxyTransport(c *rest.Config) (http.RoundTripper, error) {
	conf := &transport.Config{
		BearerToken:     c.BearerToken,
		BearerTokenFile: c.BearerTokenFile,
		TLS: transport.TLSConfig{
			Insecure: true,
		},
	}
	if c.ExecProvider != nil {
		provider, err := exec.GetAuthenticator(c.ExecProvider)
		if err != nil {
			return nil, err
		}
		if err := provider.UpdateTransportConfig(conf); err != nil {
			return nil, err
		}
	}
	if c.AuthProvider != nil {
		provider, err := rest.GetAuthProvider(c.Host, c.AuthProvider, c.AuthConfigPersister)
		if err != nil {
			return nil, err
		}
		conf.Wrap(provider.WrapTransport)
	}
	t, err := transport.New(conf)
	if err != nil {
		return nil, xerrors.Errorf("could not create a transport: %w", err)
	}
	return t, nil
}

func parseRemoteURL(r *resolver, namespace string, u *url.URL) (*v1.Pod, int, error) {
	h := u.Hostname()
	if strings.HasSuffix(h, ".svc") {
		serviceName := strings.TrimSuffix(h, ".svc")
		return r.FindPodByService(namespace, serviceName)
	}
	return r.FindContainerPort(namespace, h)
}

func newResolver(config *rest.Config) (*resolver, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, xerrors.Errorf("could not create a client: %w", err)
	}
	return &resolver{clientset.CoreV1()}, nil
}

type resolver struct {
	CoreV1 corev1.CoreV1Interface
}

func (r *resolver) FindPodByService(namespace, serviceName string) (*v1.Pod, int, error) {
	service, err := r.CoreV1.Services(namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find the service: %w", err)
	}
	var selectors []string
	for k, v := range service.Spec.Selector {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(selectors, ",")
	pods, err := r.CoreV1.Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find pods by selector %s: %w", selector, err)
	}
	if len(pods.Items) == 0 {
		return nil, 0, xerrors.Errorf("no pod matched to selector %s", selector)
	}
	pod := &pods.Items[0]
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			return pod, int(port.ContainerPort), nil
		}
	}
	return nil, 0, xerrors.Errorf("no container port in the pod %s", pod.Name)
}

func (r *resolver) FindContainerPort(namespace, podName string) (*v1.Pod, int, error) {
	pod, err := r.CoreV1.Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find the pod: %w", err)
	}
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			return pod, int(port.ContainerPort), nil
		}
	}
	return nil, 0, xerrors.Errorf("no container port in the pod %s", pod.Name)
}

func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, xerrors.Errorf("could not listen: %w", err)
	}
	defer l.Close()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, xerrors.Errorf("unknown type %T", l.Addr())
	}
	return addr.Port, nil
}
