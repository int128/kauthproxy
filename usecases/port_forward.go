package usecases

import (
	"context"
	"crypto/tls"
	"net/http"

	"gitlab.com/int128/kubectl-oidc-port-forward/portforward"
	"gitlab.com/int128/kubectl-oidc-port-forward/reverseproxy"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

type PortForwardIn struct {
	KubectlFlags   []string
	Token          string
	SourcePort     int
	TargetResource string
	TargetScheme   string
	TargetPort     int
}

func PortForward(ctx context.Context, in PortForwardIn) error {
	eg, ctx := errgroup.WithContext(ctx)

	if err := portforward.Start(ctx, eg, portforward.Source{
		Port: 8443, //TODO: allocate a free port
	}, portforward.Target{
		KubectlFlags: in.KubectlFlags,
		Resource:     in.TargetResource,
		Port:         in.TargetPort,
	}); err != nil {
		return xerrors.Errorf("could not start a kubectl process: %w", err)
	}

	modifier := func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+in.Token)
	}
	reverseproxy.Start(ctx, eg, reverseproxy.Source{
		Port: in.SourcePort,
	}, reverseproxy.Target{
		Transport: &http.Transport{
			//TODO: set timeouts
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Scheme: in.TargetScheme,
		Port:   8443, //TODO: allocate a free port
	}, modifier)

	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while port-forwarding: %w", err)
	}
	return nil
}
