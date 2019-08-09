package cmd

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/int128/kubectl-auth-port-forward/usecases"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Run(ctx context.Context, osArgs []string, version string) int {
	rootCmd := newRootCmd(ctx)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.Version = version

	rootCmd.SetArgs(osArgs[1:])
	if err := rootCmd.Execute(); err != nil {
		log.Printf("error: %s", err)
		return 1
	}
	return 0
}

type rootCmdOptions struct {
	*genericclioptions.ConfigFlags
}

func (o *rootCmdOptions) Namespace() string {
	if o.ConfigFlags.Namespace != nil && *o.ConfigFlags.Namespace != "" {
		return *o.ConfigFlags.Namespace
	}
	return "default"
}

func newRootCmd(ctx context.Context) *cobra.Command {
	var o rootCmdOptions
	o.ConfigFlags = genericclioptions.NewConfigFlags(false)
	c := &cobra.Command{
		Use:     "kubectl auth-port-forward POD_NAME LOCAL_PORT:POD_SCHEME/POD_PORT",
		Short:   "Forward a local port to a pod",
		Example: `  kubectl -n kube-system auth-port-forward kubernetes-dashboard-xxx 8443:https/8443`,
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			portPair, err := parsePortPairNotation(args[1])
			if err != nil {
				return xerrors.Errorf("invalid port pair: %w", err)
			}
			config, err := o.ConfigFlags.ToRESTConfig()
			if err != nil {
				return xerrors.Errorf("could not load the config: %w", err)
			}
			in := usecases.PortForwardIn{
				Config:             config,
				Namespace:          o.Namespace(),
				PodName:            args[0],
				PodContainerPort:   portPair.remotePort,
				PodContainerScheme: portPair.remoteScheme,
				LocalPort:          portPair.localPort,
			}
			if err := usecases.PortForward(ctx, in); err != nil {
				return xerrors.Errorf("error while port forwarding: %w", err)
			}
			return nil
		},
	}
	o.ConfigFlags.AddFlags(c.Flags())
	return c
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
