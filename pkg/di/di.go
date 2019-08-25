//+build wireinject

// Package di provides dependency injection.
package di

import (
	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/cmd"
	"github.com/int128/kauthproxy/pkg/logger"
	"github.com/int128/kauthproxy/pkg/network"
	"github.com/int128/kauthproxy/pkg/portforwarder"
	"github.com/int128/kauthproxy/pkg/resolver"
	"github.com/int128/kauthproxy/pkg/reverseproxy"
	"github.com/int128/kauthproxy/pkg/usecases"
)

func NewCmd() cmd.Interface {
	wire.Build(
		cmd.Set,
		usecases.Set,
		reverseproxy.Set,
		portforwarder.Set,
		resolver.Set,
		network.Set,
		logger.Set,
	)
	return nil
}
