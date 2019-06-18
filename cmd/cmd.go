package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Run(osArgs []string) int {
	var exitCode int
	f := genericclioptions.NewConfigFlags()
	rootCmd := cobra.Command{
		Use:     "kubectl oidc-port-forward TYPE/NAME [options] LOCAL_PORT:REMOTE_PORT",
		Short:   "Forward one or more local ports to a pod",
		Example: `  kubectl -n kube-system oidc-port-forward svc/kubernetes-dashboard 8443:443`,
		Args:    cobra.MinimumNArgs(2),
		Run: func(*cobra.Command, []string) {
			if err := runPortForward(f, osArgs); err != nil {
				log.Printf("error: %s", err)
				exitCode = 1
			}
		},
	}
	f.AddFlags(rootCmd.PersistentFlags())

	rootCmd.Version = "v0.0.1"
	rootCmd.SetArgs(osArgs[1:])
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return exitCode
}
