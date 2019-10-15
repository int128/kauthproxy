//+build wireinject

// Package di provides dependency injection.
package di

import (
	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/adaptors/cmd"
	"github.com/int128/kauthproxy/pkg/adaptors/logger"
	"github.com/int128/kauthproxy/pkg/adaptors/network"
	"github.com/int128/kauthproxy/pkg/adaptors/portforwarder"
	"github.com/int128/kauthproxy/pkg/adaptors/resolver"
	"github.com/int128/kauthproxy/pkg/adaptors/reverseproxy"
	"github.com/int128/kauthproxy/pkg/usecases/authproxy"
)

func NewCmd() cmd.Interface {
	wire.Build(
		// adaptors
		cmd.Set,
		reverseproxy.Set,
		portforwarder.Set,
		resolver.Set,
		network.Set,
		logger.Set,

		// usecases
		authproxy.Set,
	)
	return nil
}
