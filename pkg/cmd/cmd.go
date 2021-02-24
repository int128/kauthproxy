// Package cmd provides command line interface.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/authproxy"
	"github.com/int128/kauthproxy/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var Set = wire.NewSet(
	wire.Struct(new(Cmd), "*"),
	wire.Bind(new(Interface), new(*Cmd)),
)

type Interface interface {
	Run(ctx context.Context, osArgs []string, version string) int
}

var defaultAddress = []string{
	"127.0.0.1:18000",
	"127.0.0.1:28000",
}

// Cmd provides command line interface.
type Cmd struct {
	AuthProxy authproxy.Interface
	Logger    logger.Interface
}

// Run parses the arguments and executes the corresponding use-case.
func (cmd *Cmd) Run(ctx context.Context, osArgs []string, version string) int {
	rootCmd := cmd.newRootCmd()
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.Version = version

	rootCmd.SetArgs(osArgs[1:])
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			cmd.Logger.V(1).Infof("terminating: %s", err)
			return 0
		}
		cmd.Logger.Printf("error: %s", err)
		cmd.Logger.V(1).Infof("stacktrace: %+v", err)
		return 1
	}
	return 0
}

type rootCmdOptions struct {
	k8sOptions        *genericclioptions.ConfigFlags
	addressCandidates []string
	skipOpenBrowser   bool
}

func (o *rootCmdOptions) addFlags(f *pflag.FlagSet) {
	o.k8sOptions.AddFlags(f)
	f.StringArrayVar(&o.addressCandidates, "address", defaultAddress, "The address on which to run the proxy. If set multiple times, it will try binding the address in order")
	f.BoolVar(&o.skipOpenBrowser, "skip-open-browser", false, "If set, skip opening the browser")
}

func (cmd *Cmd) newRootCmd() *cobra.Command {
	var o rootCmdOptions
	o.k8sOptions = genericclioptions.NewConfigFlags(false)
	c := &cobra.Command{
		Use:   "kubectl auth-proxy POD_OR_SERVICE_URL",
		Short: "Forward a local port to a pod or service via the authentication proxy",
		Long: `Forward a local port to a pod or service via the authentication proxy.
It gets a token from the current credential plugin (e.g. EKS, OpenID Connect).
Then it appends the authorization header to HTTP requests, like "authorization: Bearer token".
All traffic is routed by the authentication proxy and port forwarder as follows:
  [browser] -> [authentication proxy] -> [port forwarder] -> [pod]`,
		Example: `  # To access a service:
  kubectl auth-proxy https://kubernetes-dashboard.svc

  # To access a pod:
  kubectl auth-proxy https://kubernetes-dashboard-57fc4fcb74-jjg77`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.runRootCmd(c.Context(), o, args)
		},
	}
	o.addFlags(c.Flags())
	cmd.Logger.AddFlags(c.PersistentFlags())
	return c
}

func (cmd *Cmd) runRootCmd(ctx context.Context, o rootCmdOptions, args []string) error {
	remoteURL, err := url.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid remote URL: %w", err)
	}
	config, err := o.k8sOptions.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("could not load the config: %w", err)
	}
	namespace, _, err := o.k8sOptions.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return fmt.Errorf("could not determine the namespace: %w", err)
	}
	authProxyOption := authproxy.Option{
		Config:                config,
		Namespace:             namespace,
		TargetURL:             remoteURL,
		BindAddressCandidates: o.addressCandidates,
		SkipOpenBrowser:       o.skipOpenBrowser,
	}
	if err := cmd.AuthProxy.Do(ctx, authProxyOption); err != nil {
		return fmt.Errorf("could not run an authentication proxy: %w", err)
	}
	return nil
}
