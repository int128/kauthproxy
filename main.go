package main

import (
	"context"
	"os"
	"os/signal"

	"gitlab.com/int128/kubectl-oidc-port-forward/cmd"
)

var version = "v0.0.0"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// cancel the context on interrupted (ctrl+c)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)
	go func() {
		<-signals
		cancel()
	}()

	os.Exit(cmd.Run(ctx, os.Args, version))
}
