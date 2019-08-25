// Package reverseproxy provides a reverse proxy.
package reverseproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/logger"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

var Set = wire.NewSet(
	wire.Struct(new(ReverseProxy), "*"),
	wire.Bind(new(Interface), new(*ReverseProxy)),
)

//go:generate mockgen -destination mock_reverseproxy/mock_reverseproxy.go github.com/int128/kauthproxy/pkg/reverseproxy Interface

type Interface interface {
	Start(ctx context.Context, eg *errgroup.Group, o Options)
}

// ReverseProxy provides a reverse proxy.
type ReverseProxy struct {
	Logger logger.Interface
}

// Options represents an option of ReverseProxy.
type Options struct {
	Transport http.RoundTripper
	Source    Source
	Target    Target
}

// Source represents a source of proxy.
type Source struct {
	Address string // local address to bind
}

// Target represents a target of proxy.
type Target struct {
	Scheme string
	Host   string
	Port   int
}

// Start starts a reverse proxy in goroutines.
func (rp *ReverseProxy) Start(ctx context.Context, eg *errgroup.Group, o Options) {
	server := &http.Server{
		Addr: o.Source.Address,
		Handler: &httputil.ReverseProxy{
			Transport: o.Transport,
			Director: func(r *http.Request) {
				r.URL.Scheme = o.Target.Scheme
				r.URL.Host = fmt.Sprintf("%s:%d", o.Target.Host, o.Target.Port)
				r.Host = ""
			},
		},
	}
	eg.Go(func() error {
		rp.Logger.V(1).Infof("starting a reverse proxy for %s -> %s:%d", o.Source.Address, o.Target.Host, o.Target.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return xerrors.Errorf("could not start a reverse proxy: %w", err)
		}
		rp.Logger.V(1).Infof("stopped the reverse proxy")
		return nil
	})
	eg.Go(func() error {
		<-ctx.Done()
		rp.Logger.V(1).Infof("stopping the reverse proxy")
		if err := server.Shutdown(ctx); err != nil {
			return xerrors.Errorf("could not stop the reverse proxy: %w", err)
		}
		return nil
	})
}
