package portforward

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

type Source struct {
	Port int
}

type Target struct {
	KubectlFlags []string
	Resource     string // TYPE/NAME, e.g. svc/kubernetes-dashboard
	Port         int
}

func Start(ctx context.Context, eg *errgroup.Group, source Source, target Target) error {
	args := []string{"port-forward"}
	args = append(args, target.KubectlFlags...)
	args = append(args, target.Resource)
	args = append(args, fmt.Sprintf("%d:%d", source.Port, target.Port))

	log.Printf("Starting kubectl %v", args)
	c := exec.Command("kubectl", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return xerrors.Errorf("could not run kubectl %+v: %w", args, err)
	}
	log.Printf("kubectl running")
	eg.Go(func() error {
		if err := c.Wait(); err != nil {
			return xerrors.Errorf("error while running kubectl %+v: %w", args, err)
		}
		log.Printf("kubectl exited")
		return nil
	})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			log.Printf("sending a signal to kubectl process")
			if err := c.Process.Signal(os.Interrupt); err != nil {
				return xerrors.Errorf("error while sending a signal to kubectl process: %w", err)
			}
		}
		return nil
	})
	return nil
}
