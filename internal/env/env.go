package env

import (
	"log"
	"fmt"
	"net"

	"github.com/google/wire"
)

var Set = wire.NewSet(
	wire.Struct(new(Env), "*"),
	wire.Bind(new(Interface), new(*Env)),
)

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
	defer func() {
		if err := l.Close(); err != nil {
			log.Printf("Failed to close listener: %v", err)
		}
	}()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("internal error: unknown type %T", l.Addr())
	}
	return addr.Port, nil
}
