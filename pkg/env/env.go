package env

import (
	"fmt"
	"net"

	"github.com/google/wire"
)

var Set = wire.NewSet(
	wire.Struct(new(Env), "*"),
	wire.Bind(new(Interface), new(*Env)),
)

//go:generate mockgen -destination mock_env/mock_env.go github.com/int128/kauthproxy/pkg/env Interface

type Interface interface {
	AllocateLocalPort() (int, error)
}

type Env struct{}

// AllocateLocalPort returns a free port on localhost.
func (*Env) AllocateLocalPort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("could not listen: %w", err)
	}
	defer l.Close()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("internal error: unknown type %T", l.Addr())
	}
	return addr.Port, nil
}
