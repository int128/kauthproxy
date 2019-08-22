package usecases

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/int128/kauthproxy/pkg/reverseproxy"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/plugin/pkg/client/auth/exec"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardIn struct {
	Config    *rest.Config
	Namespace string
	RemoteURL *url.URL
	LocalAddr string
}

func PortForward(ctx context.Context, in PortForwardIn) error {
	cfg := in.Config

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return xerrors.Errorf("could not create a client: %w", err)
	}

	pod, containerPort, err := resolvePodContainerPort(in.RemoteURL, clientset, in.Namespace)
	if err != nil {
		return xerrors.Errorf("could not resolve a pod: %w", err)
	}
	log.Printf("Pod %s is %s", pod.Name, pod.Status.Phase)
	portforwardURL, err := url.Parse(cfg.Host + pod.GetSelfLink() + "/portforward")
	if err != nil {
		return xerrors.Errorf("could not create a URL for portforward: %w", err)
	}
	log.Printf("Forwarder URL: %s", portforwardURL)

	spdyTransport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return xerrors.Errorf("could not create a round tripper: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: spdyTransport}, http.MethodPost, portforwardURL)

	transitPort, err := findFreePort()
	if err != nil {
		return xerrors.Errorf("could not allocate a local port: %w", err)
	}
	portNotation := fmt.Sprintf("%d:%s", transitPort, containerPort)
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{})
	forwarder, err := portforward.New(dialer, []string{portNotation}, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return xerrors.Errorf("could not create a forwarder: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	proxyTransport, err := newProxyTransport(in.Config)
	if err != nil {
		return xerrors.Errorf("could not create a transport for reverse proxy: %w", err)
	}
	reverseproxy.Start(ctx, eg,
		reverseproxy.Source{Addr: in.LocalAddr},
		reverseproxy.Target{
			Transport: proxyTransport,
			Scheme:    in.RemoteURL.Scheme,
			Port:      transitPort,
		}, func(r *http.Request) {})
	go func() {
		<-ctx.Done()
		close(stopChan)
	}()
	eg.Go(func() error {
		if err := forwarder.ForwardPorts(); err != nil {
			return xerrors.Errorf("error while running the forwarder: %w", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while port-forwarding: %w", err)
	}
	return nil
}

func newProxyTransport(c *rest.Config) (http.RoundTripper, error) {
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

func resolvePodContainerPort(url *url.URL, clientset *kubernetes.Clientset, namespace string) (*v1.Pod, string, error) {
	hostname := url.Hostname()
	var pod *v1.Pod
	if strings.HasSuffix(hostname, ".svc") {
		serviceName := strings.TrimSuffix(hostname, ".svc")
		service, err := clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
		if err != nil {
			return nil, "", xerrors.Errorf("could not find the service: %w", err)
		}
		log.Printf("Service %s found", service.Name)
		var selectors []string
		for k, v := range service.Spec.Selector {
			selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
		}
		selector := strings.Join(selectors, ",")
		pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return nil, "", xerrors.Errorf("could not find the pods by selector %s: %w", selector, err)
		}
		if len(pods.Items) == 0 {
			return nil, "", xerrors.Errorf("no pod matched to selector %s", selector)
		}
		pod = &pods.Items[0]
	} else {
		var err error
		pod, err = clientset.CoreV1().Pods(namespace).Get(hostname, metav1.GetOptions{})
		if err != nil {
			return nil, "", xerrors.Errorf("could not find the pod: %w", err)
		}
	}
	if url.Port() != "" {
		return pod, url.Port(), nil
	}
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			return pod, fmt.Sprintf("%d", port.ContainerPort), nil
		}
	}
	return nil, "", xerrors.Errorf("no container port found in the pod %s", pod.Name)
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
