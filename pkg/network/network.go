package network

import (
	"net"
	"net/http"

	"github.com/google/wire"
	"golang.org/x/xerrors"
	"k8s.io/client-go/plugin/pkg/client/auth/exec"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

var Set = wire.NewSet(
	wire.Struct(new(Network), "*"),
	wire.Bind(new(Interface), new(*Network)),
)

type Interface interface {
	NewTransportWithToken(c *rest.Config) (http.RoundTripper, error)
	AllocateLocalPort() (int, error)
}

type Network struct{}

// NewTransportWithToken returns a RoundTripper with token support.
func (*Network) NewTransportWithToken(c *rest.Config) (http.RoundTripper, error) {
	config := &transport.Config{
		BearerToken:     c.BearerToken,
		BearerTokenFile: c.BearerTokenFile,
		TLS: transport.TLSConfig{
			Insecure: true,
		},
	}
	if c.ExecProvider != nil {
		provider, err := exec.GetAuthenticator(c.ExecProvider)
		if err != nil {
			return nil, xerrors.Errorf("could not get an authenticator: %w", err)
		}
		if err := provider.UpdateTransportConfig(config); err != nil {
			return nil, xerrors.Errorf("could not update the transport config: %w", err)
		}
	}
	if c.AuthProvider != nil {
		provider, err := rest.GetAuthProvider(c.Host, c.AuthProvider, c.AuthConfigPersister)
		if err != nil {
			return nil, xerrors.Errorf("could not get an auth-provider: %w", err)
		}
		config.Wrap(provider.WrapTransport)
	}
	t, err := transport.New(config)
	if err != nil {
		return nil, xerrors.Errorf("could not create a transport: %w", err)
	}
	return t, nil
}

// AllocateLocalPort returns a free port on localhost.
func (*Network) AllocateLocalPort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, xerrors.Errorf("could not listen: %w", err)
	}
	defer l.Close()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, xerrors.Errorf("internal error: unknown type %T", l.Addr())
	}
	return addr.Port, nil
}
