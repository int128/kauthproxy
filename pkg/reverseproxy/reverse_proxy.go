package reverseproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/google/wire"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

var Set = wire.NewSet(
	wire.Struct(new(ReverseProxy), "*"),
	wire.Bind(new(Interface), new(*ReverseProxy)),
)

type Interface interface {
	Start(ctx context.Context, eg *errgroup.Group, local Local, remote Remote)
}

type ReverseProxy struct {
}

type Local struct {
	Addr string
}

type Remote struct {
	Transport http.RoundTripper
	Scheme    string
	Port      int
}

func (*ReverseProxy) Start(ctx context.Context, eg *errgroup.Group, local Local, remote Remote) {
	server := &http.Server{
		Addr: local.Addr,
		Handler: &httputil.ReverseProxy{
			Transport: remote.Transport,
			Director: func(r *http.Request) {
				r.URL.Scheme = remote.Scheme
				r.URL.Host = fmt.Sprintf("localhost:%d", remote.Port)
				r.Host = ""
			},
		},
	}
	eg.Go(func() error {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return xerrors.Errorf("could not start a server: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		<-ctx.Done()
		log.Printf("Shutting down the server")
		if err := server.Shutdown(ctx); err != nil {
			return xerrors.Errorf("could not stop the server: %w", err)
		}
		return nil
	})
}
