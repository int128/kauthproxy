// Package reverseproxy provides a reverse proxy.
package reverseproxy

import (
	"context"
	"fmt"
	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/logger"
	"github.com/int128/listener"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"net/http"
	"net/http/httputil"
)

var Set = wire.NewSet(
	wire.Struct(new(ReverseProxy), "*"),
	wire.Bind(new(Interface), new(*ReverseProxy)),
)

//go:generate mockgen -destination mock_reverseproxy/mock_reverseproxy.go github.com/int128/kauthproxy/pkg/reverseproxy Interface

// Options represents an option of a reverse proxy.
type Options struct {
	Transport http.RoundTripper
	Source    Source
	Target    Target
}

// Source represents a source of proxy.
type Source struct {
	AddressCandidates []string
}

// Target represents a target of proxy.
type Target struct {
	Scheme string
	Host   string
	Port   int
}

type Interface interface {
	Run(ctx context.Context, o Options) error
}

type ReverseProxy struct {
	Logger logger.Interface
}

// Run starts a server and waits until the context is canceled.
// In most case it returns an error which wraps context.Canceled.
func (rp *ReverseProxy) Run(ctx context.Context, o Options) error {
	s := &http.Server{
		Handler: &httputil.ReverseProxy{
			Transport: o.Transport,
			Director: func(r *http.Request) {
				r.URL.Scheme = o.Target.Scheme
				r.URL.Host = fmt.Sprintf("%s:%d", o.Target.Host, o.Target.Port)
				r.Host = ""
			},
		},
	}
	l, err := listener.New(o.Source.AddressCandidates)
	if err != nil {
		return xerrors.Errorf("could not listen: %w", err)
	}
	// l will be closed by s.Serve(l)

	rp.Logger.Printf("Open %s", l.URL)

	finalizeChan := make(chan struct{})
	var eg errgroup.Group
	eg.Go(func() error {
		defer close(finalizeChan)
		rp.Logger.V(1).Infof("starting a server at %s", l.Addr().String())
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			return xerrors.Errorf("could not start a server: %w", err)
		}
		rp.Logger.V(1).Infof("stopped the server at %s", l.Addr().String())
		return nil
	})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			rp.Logger.V(1).Infof("stopping the server at %s", l.Addr().String())
			if err := s.Shutdown(ctx); err != nil {
				return xerrors.Errorf("could not stop the server at %s: %w", l.Addr().String(), err)
			}
			return xerrors.Errorf("stopping the server: %w", ctx.Err())
		case <-finalizeChan:
			rp.Logger.V(1).Infof("finished goroutine of the server at %s", l.Addr().String())
			return nil
		}
	})
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while running a reverse proxy: %w", err)
	}
	return nil
}
