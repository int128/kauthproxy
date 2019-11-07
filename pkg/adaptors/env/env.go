package env

import (
	"net"

	"github.com/google/wire"
	"github.com/pkg/browser"
	"golang.org/x/xerrors"
)

var Set = wire.NewSet(
	wire.Struct(new(Env), "*"),
	wire.Bind(new(Interface), new(*Env)),
)

//go:generate mockgen -destination mock_env/mock_env.go github.com/int128/kauthproxy/pkg/adaptors/env Interface

type Interface interface {
	AllocateLocalPort() (int, error)
	OpenBrowser(url string) error
}

type Env struct{}

// AllocateLocalPort returns a free port on localhost.
func (*Env) AllocateLocalPort() (int, error) {
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

// OpenBrowser opens the default browser.
func (*Env) OpenBrowser(url string) error {
	if err := browser.OpenURL(url); err != nil {
		return xerrors.Errorf("could not open the browser: %w", err)
	}
	return nil
}
