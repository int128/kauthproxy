package cmd

import (
	"context"
	"log"
	"net"
	"net/url"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/usecases"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var Set = wire.NewSet(
	wire.Struct(new(Cmd), "*"),
	wire.Bind(new(Interface), new(*Cmd)),
)

type Interface interface {
	Run(ctx context.Context, osArgs []string, version string) int
}

type Cmd struct {
	PortForward usecases.PortForwardInterface
}

func (cmd *Cmd) Run(ctx context.Context, osArgs []string, version string) int {
	rootCmd := cmd.newRootCmd(ctx)
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

func (cmd *Cmd) newRootCmd(ctx context.Context) *cobra.Command {
	var o rootCmdOptions
	o.ConfigFlags = genericclioptions.NewConfigFlags(false)
	c := &cobra.Command{
		Use:   "kubectl auth-proxy REMOTE_URL [LOCAL_ADDR]",
		Short: "Forward a local port to a pod or service via authentication proxy",
		Long: `Forward a local port to a pod or service via authentication proxy.

To forward a local port to a service, set a service name with .svc suffix. e.g. http://service-name.svc
To forward a local port to a pod, set a pod name. e.g. http://pod-name

LOCAL_ADDR defaults to localhost:8000.
`,
		Example: `  kubectl auth-proxy https://kubernetes-dashboard.svc`,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.runRootCmd(ctx, o, args)
		},
	}
	o.ConfigFlags.AddFlags(c.Flags())
	return c
}

func (cmd *Cmd) runRootCmd(ctx context.Context, o rootCmdOptions, args []string) error {
	remoteURL, err := url.Parse(args[0])
	if err != nil {
		return xerrors.Errorf("invalid remote URL: %w", err)
	}
	localAddr := "localhost:8000"
	if len(args) == 2 {
		if _, _, err := net.SplitHostPort(args[1]); err != nil {
			return xerrors.Errorf("invalid local address: %w", err)
		}
		localAddr = args[1]
	}
	config, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return xerrors.Errorf("could not load the config: %w", err)
	}
	namespace, _, err := o.ConfigFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return xerrors.Errorf("could not determine the namespace: %w", err)
	}
	in := usecases.PortForwardIn{
		Config:    config,
		Namespace: namespace,
		RemoteURL: remoteURL,
		LocalAddr: localAddr,
	}
	if err := cmd.PortForward.Do(ctx, in); err != nil {
		return xerrors.Errorf("error while port forwarding: %w", err)
	}
	return nil
}
