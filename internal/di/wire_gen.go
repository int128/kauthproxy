// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
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

// Injectors from di.go:

func NewCmd() cmd.Interface {
	reverseProxy := &reverseproxy.ReverseProxy{}
	portForwarder := &portforwarder.PortForwarder{}
	loggerLogger := &logger.Logger{}
	factory := &resolver.Factory{
		Logger: loggerLogger,
	}
	newFunc := _wireNewFuncValue
	envEnv := &env.Env{}
	browserBrowser := &browser.Browser{}
	authProxy := &authproxy.AuthProxy{
		ReverseProxy:    reverseProxy,
		PortForwarder:   portForwarder,
		ResolverFactory: factory,
		NewTransport:    newFunc,
		Env:             envEnv,
		Browser:         browserBrowser,
		Logger:          loggerLogger,
	}
	cmdCmd := &cmd.Cmd{
		AuthProxy: authProxy,
		Logger:    loggerLogger,
	}
	return cmdCmd
}

var (
	_wireNewFuncValue = transport.NewFunc(transport.New)
)
