package portforward

import (
	"context"
	"log"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

func Start(ctx context.Context, eg *errgroup.Group, args []string) error {
	args = append([]string{"port-forward"}, args...)
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
