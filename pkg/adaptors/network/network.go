package network

import (
	"net"

	"github.com/google/wire"
	"golang.org/x/xerrors"
)

var Set = wire.NewSet(
	wire.Struct(new(Network), "*"),
	wire.Bind(new(Interface), new(*Network)),
)

//go:generate mockgen -destination mock_network/mock_network.go github.com/int128/kauthproxy/pkg/adaptors/network Interface

type Interface interface {
	AllocateLocalPort() (int, error)
}

type Network struct{}

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
