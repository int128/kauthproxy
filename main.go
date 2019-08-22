package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/int128/kauthproxy/pkg/cmd"
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
