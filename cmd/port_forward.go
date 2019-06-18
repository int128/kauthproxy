package cmd

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"

	"gitlab.com/int128/kubectl-oidc-port-forward/portforward"
	"gitlab.com/int128/kubectl-oidc-port-forward/reverseproxy"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func runPortForward(f *genericclioptions.ConfigFlags, osArgs []string) error {
	config, err := f.ToRESTConfig()
	if err != nil {
		return xerrors.Errorf("could not load the config: %w", err)
	}
	token := config.AuthProvider.Config["id-token"]
	modifier := func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+token)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)
	go func() {
		<-signals
		cancel()
	}()

	eg, ctx := errgroup.WithContext(ctx)
	if err := portforward.Start(ctx, eg, osArgs[1:]); err != nil {
		return xerrors.Errorf("could not start a kubectl process: %w", err)
	}
	reverseproxy.Start(ctx, eg, 8888, reverseproxy.Target{
		Transport: &http.Transport{
			//TODO: set timeouts
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Scheme: "https",
		Port:   8443,
	}, modifier)
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while port-forwarding: %w", err)
	}
	return nil
}
