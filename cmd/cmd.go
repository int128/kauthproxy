package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gitlab.com/int128/kubectl-oidc-port-forward/usecases"
	"golang.org/x/xerrors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Run(ctx context.Context, osArgs []string, version string) int {
	var exitCode int
	f := genericclioptions.NewConfigFlags()
	rootCmd := cobra.Command{
		Use:     "kubectl oidc-port-forward TYPE/NAME [options] LOCAL_PORT:SCHEME/REMOTE_PORT",
		Short:   "Forward one or more local ports to a pod",
		Example: `  kubectl -n kube-system oidc-port-forward svc/kubernetes-dashboard 8443:https/443`,
		Args:    cobra.MinimumNArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			if err := runPortForward(ctx, f, args, osArgs); err != nil {
				log.Printf("error: %s", err)
				exitCode = 1
			}
		},
	}
	f.AddFlags(rootCmd.Flags())

	rootCmd.Version = version
	rootCmd.SetArgs(osArgs[1:])
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return exitCode
}

func runPortForward(ctx context.Context, f *genericclioptions.ConfigFlags, args, osArgs []string) error {
	config, err := f.ToRESTConfig()
	if err != nil {
		return xerrors.Errorf("could not load the config: %w", err)
	}
	token := config.AuthProvider.Config["id-token"]

	kubectlFlags, err := extractKubectlFlags(osArgs)
	if err != nil {
		return xerrors.Errorf("could not extract the kubectl flags: %w", err)
	}

	if err := usecases.PortForward(ctx, usecases.PortForwardIn{
		KubectlFlags:   kubectlFlags,
		Token:          token,
		SourcePort:     8888,                       //TODO: parse args
		TargetResource: "svc/kubernetes-dashboard", //TODO: parse args
		TargetScheme:   "https",                    //TODO: parse args
		TargetPort:     443,                        //TODO: parse args
	}); err != nil {
		return xerrors.Errorf("error while port forwarding: %w", err)
	}
	return nil
}

func extractKubectlFlags(osArgs []string) ([]string, error) {
	f := genericclioptions.NewConfigFlags()
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	f.AddFlags(fs)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	if err := fs.Parse(osArgs[1:]); err != nil {
		return nil, xerrors.Errorf("could not parse the arguments: %w", err)
	}
	var flags []string
	fs.Visit(func(f *pflag.Flag) {
		flags = append(flags, "--"+f.Name, f.Value.String())
	})
	return flags, nil
}
