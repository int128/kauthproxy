package cmd

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/int128/kubectl-auth-port-forward/usecases"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Run(ctx context.Context, osArgs []string, version string) int {
	var exitCode int

	rootOpt := &rootCmdOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(false),
		osArgs:      osArgs,
	}
	rootCmd := cobra.Command{
		Use:     "kubectl auth-port-forward TYPE/NAME [options] LOCAL_PORT:REMOTE_SCHEME/REMOTE_PORT",
		Short:   "Forward a local port to a pod",
		Example: `  kubectl -n kube-system auth-port-forward svc/kubernetes-dashboard 8443:https/443`,
		Args:    cobra.ExactArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			rootOpt.cmdArgs = args
			if err := runRootCmd(ctx, rootOpt); err != nil {
				log.Printf("error: %s", err)
				exitCode = 1
			}
		},
	}
	rootOpt.ConfigFlags.AddFlags(rootCmd.Flags())

	rootCmd.Version = version
	rootCmd.SetArgs(osArgs[1:])
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return exitCode
}

type rootCmdOptions struct {
	*genericclioptions.ConfigFlags
	osArgs  []string
	cmdArgs []string
}

func runRootCmd(ctx context.Context, o *rootCmdOptions) error {
	config, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return xerrors.Errorf("could not load the config: %w", err)
	}
	token := config.AuthProvider.Config["id-token"]
	if token == "" {
		return xerrors.Errorf("could not find a token from the kubeconfig")
	}

	kubectlFlags, err := extractKubectlFlags(o.osArgs)
	if err != nil {
		return xerrors.Errorf("could not extract the kubectl flags: %w", err)
	}

	targetResource, portPairNotation := o.cmdArgs[0], o.cmdArgs[1]
	pair, err := parsePortPairNotation(portPairNotation)
	if err != nil {
		return xerrors.Errorf("invalid port pair notation `%s`: %w", portPairNotation, err)
	}
	if err := usecases.PortForward(ctx, usecases.PortForwardIn{
		KubectlFlags:   kubectlFlags,
		Token:          token,
		SourcePort:     pair.localPort,
		TargetPort:     pair.remotePort,
		TargetScheme:   pair.remoteScheme,
		TargetResource: targetResource,
	}); err != nil {
		return xerrors.Errorf("error while port forwarding: %w", err)
	}
	return nil
}

type portPair struct {
	localPort    int
	remotePort   int
	remoteScheme string
}

func parsePortPairNotation(s string) (*portPair, error) {
	localRemotePair := strings.SplitN(s, ":", 2)
	if len(localRemotePair) != 2 {
		return nil, xerrors.Errorf("notation must contain a colon")
	}
	localPortString, remoteSchemePort := localRemotePair[0], localRemotePair[1]
	localPort, err := strconv.Atoi(localPortString)
	if err != nil {
		return nil, xerrors.Errorf("invalid local port: %w", err)
	}
	remoteSchemePortPair := strings.SplitN(remoteSchemePort, "/", 2)
	if len(remoteSchemePortPair) != 2 {
		return nil, xerrors.Errorf("remote notation must contain a slash")
	}
	remoteScheme, remotePortString := remoteSchemePortPair[0], remoteSchemePortPair[1]
	remotePort, err := strconv.Atoi(remotePortString)
	if err != nil {
		return nil, xerrors.Errorf("invalid remote port: %w", err)
	}
	return &portPair{
		localPort:    localPort,
		remotePort:   remotePort,
		remoteScheme: remoteScheme,
	}, nil
}

func extractKubectlFlags(osArgs []string) ([]string, error) {
	f := genericclioptions.NewConfigFlags(false)
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
