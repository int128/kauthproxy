// Package reverseproxy provides a reverse proxy.
package reverseproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/google/wire"
	"github.com/int128/listener"
)

var Set = wire.NewSet(
	wire.Struct(new(ReverseProxy), "*"),
	wire.Bind(new(Interface), new(*ReverseProxy)),
)

//go:generate mockgen -destination mock_reverseproxy/mock_reverseproxy.go github.com/int128/kauthproxy/internal/reverseproxy Interface,Instance

// Option represents an option of a reverse proxy.
type Option struct {
	Transport             http.RoundTripper
	BindAddressCandidates []string
	TargetScheme          string
	TargetHost            string
	TargetPort            int
}

type Interface interface {
	Run(o Option, readyChan chan<- Instance) error
}

type Instance interface {
	URL() *url.URL
	Shutdown(ctx context.Context) error
}

type ReverseProxy struct {
}

// Run executes a reverse proxy server.
//
// It returns nil if the server has been closed.
// It returns an error otherwise.
//
// It will send the Instance to the readyChan when the reverse proxy is ready.
// Caller should close the readyChan.
func (rp *ReverseProxy) Run(o Option, readyChan chan<- Instance) error {
	s := &http.Server{
		Handler: &httputil.ReverseProxy{
			Transport: o.Transport,
			Director: func(r *http.Request) {
				r.URL.Scheme = o.TargetScheme
				r.URL.Host = fmt.Sprintf("%s:%d", o.TargetHost, o.TargetPort)
				r.Host = ""
			},
		},
	}
	l, err := listener.New(o.BindAddressCandidates)
	if err != nil {
		return fmt.Errorf("could not listen: %w", err)
	}
	// l will be closed by s.Serve(l)

	if readyChan != nil {
		readyChan <- &instance{s: s, l: l}
	}
	if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not start a server: %w", err)
	}
	return nil
}

type instance struct {
	s *http.Server
	l *listener.Listener
}

func (i *instance) URL() *url.URL {
	return i.l.URL
}

func (i *instance) Shutdown(ctx context.Context) error {
	return i.s.Shutdown(ctx)
}
