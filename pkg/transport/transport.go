// Package transport provides a HTTP transport with a token got from the credential plugin of the cluster.
package transport

import (
	"net/http"

	"github.com/google/wire"
	"golang.org/x/xerrors"
	"k8s.io/client-go/plugin/pkg/client/auth/exec"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

var Set = wire.NewSet(
	wire.Value(NewFunc(New)),
)

type NewFunc func(*rest.Config) (http.RoundTripper, error)

// New returns a RoundTripper with token support.
func New(c *rest.Config) (http.RoundTripper, error) {
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
