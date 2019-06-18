package main

import (
	"gitlab.com/int128/kubectl-oidc-port-forward/cmd"
	"os"
	// https://github.com/kubernetes/client-go/issues/345
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func main() {
	os.Exit(cmd.Run(os.Args))
}
