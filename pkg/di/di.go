//+build wireinject

// Package di provides dependency injection.
package di

import (
	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/authproxy"
	"github.com/int128/kauthproxy/pkg/browser"
	"github.com/int128/kauthproxy/pkg/cmd"
	"github.com/int128/kauthproxy/pkg/env"
	"github.com/int128/kauthproxy/pkg/logger"
	"github.com/int128/kauthproxy/pkg/portforwarder"
	"github.com/int128/kauthproxy/pkg/resolver"
	"github.com/int128/kauthproxy/pkg/reverseproxy"
	"github.com/int128/kauthproxy/pkg/transport"
)

func NewCmd() cmd.Interface {
	wire.Build(
		// adaptors
		cmd.Set,
		reverseproxy.Set,
		portforwarder.Set,
		resolver.Set,
		transport.Set,
		env.Set,
		browser.Set,
		logger.Set,

		// usecases
		authproxy.Set,
	)
	return nil
}
