package reverseproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

type Source struct {
	Addr string
}

type Target struct {
	Transport http.RoundTripper
	Scheme    string
	Port      int
}

type Modifier func(r *http.Request)

func Start(ctx context.Context, eg *errgroup.Group, source Source, target Target, modifier Modifier) {
	server := &http.Server{
		Addr: source.Addr,
		Handler: &httputil.ReverseProxy{
			Transport: target.Transport,
			Director: func(r *http.Request) {
				r.URL.Scheme = target.Scheme
				r.URL.Host = fmt.Sprintf("localhost:%d", target.Port)
				r.Host = ""
				modifier(r)
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

type ReverseProxy struct {
}
