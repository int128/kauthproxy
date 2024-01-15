//go:build wireinject
// +build wireinject

// Package di provides dependency injection.
package di

import (
	"github.com/google/wire"
	"github.com/int128/kauthproxy/internal/authproxy"
	"github.com/int128/kauthproxy/internal/browser"
	"github.com/int128/kauthproxy/internal/cmd"
	"github.com/int128/kauthproxy/internal/env"
	"github.com/int128/kauthproxy/internal/logger"
	"github.com/int128/kauthproxy/internal/portforwarder"
	"github.com/int128/kauthproxy/internal/resolver"
	"github.com/int128/kauthproxy/internal/reverseproxy"
	"github.com/int128/kauthproxy/internal/transport"
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
