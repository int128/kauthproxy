//+build wireinject

// Package di provides dependency injection.
package di

import (
	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/cmd"
	"github.com/int128/kauthproxy/pkg/reverseproxy"
	"github.com/int128/kauthproxy/pkg/usecases"
)

func NewCmd() cmd.Interface {
	wire.Build(
		cmd.Set,
		usecases.Set,
		reverseproxy.Set,
	)
	return nil
}
