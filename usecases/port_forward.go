package usecases

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"gitlab.com/int128/kubectl-auth-port-forward/portforward"
	"gitlab.com/int128/kubectl-auth-port-forward/reverseproxy"
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

	transitPort, err := findFreePort()
	if err != nil {
		return xerrors.Errorf("could not find a free port: %w", err)
	}

	if err := portforward.Start(ctx, eg, portforward.Source{
		Port: transitPort,
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
		Port:   transitPort,
	}, modifier)

	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while port-forwarding: %w", err)
	}
	return nil
}

func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, xerrors.Errorf("could not listen: %w", err)
	}
	defer l.Close()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, xerrors.Errorf("unknown type %T", l.Addr())
	}
	return addr.Port, nil
}
