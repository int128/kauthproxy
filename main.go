package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/int128/kauthproxy/pkg/di"

	// for an AKS cluster integrated with Azure AD
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
)

var version = "v0.0.0"

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	os.Exit(di.NewCmd().Run(ctx, os.Args, version))
}
