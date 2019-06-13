package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	// https://github.com/kubernetes/client-go/issues/345
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type proxyHandler struct {
	baseURL string
	token   string
	client  *http.Client
}

func (h *proxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	proxyReq, err := http.NewRequest(req.Method, h.baseURL+req.URL.String(), req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	// TODO: exclude hop-by-hop headers
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers#hbh
	for k, v := range req.Header {
		for _, v := range v {
			proxyReq.Header.Add(k, v)
			log.Printf("Request header: %s: %s", k, v)
		}
	}
	proxyReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		for _, v := range v {
			w.Header().Add(k, v)
			log.Printf("Response header: %s: %s", k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("error while writing response body: %s", err)
	}
}

func execKubectl(ctx context.Context, args []string) error {
	args = append([]string{"port-forward"}, args...)
	log.Printf("Starting kubectl %v", args)
	c := exec.Command("kubectl", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return xerrors.Errorf("could not run kubectl %+v: %w", args, err)
	}
	log.Printf("kubectl running")
	go func() {
		select {
		case <-ctx.Done():
			log.Printf("sending signal to kubectl")
			if err := c.Process.Signal(os.Interrupt); err != nil {
				log.Printf("error while sending signal to kubectl: %s", err)
			}
		}
	}()
	if err := c.Wait(); err != nil {
		return xerrors.Errorf("error while running kubectl %+v: %w", args, err)
	}
	log.Printf("kubectl exited")
	return nil
}

func runReverseProxyServer(ctx context.Context, f *genericclioptions.ConfigFlags, portForwardArgs []string) error {
	config, err := f.ToRESTConfig()
	if err != nil {
		return xerrors.Errorf("could not load the config: %w", err)
	}
	token := config.AuthProvider.Config["id-token"]
	log.Printf("Using bearer token: %s", token)
	client := &http.Client{
		//TODO: set timeouts
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	server := &http.Server{
		Addr: "localhost:8888",
		Handler: &proxyHandler{
			baseURL: "https://localhost:8443",
			token:   token,
			client:  client,
		},
	}
	log.Printf("Open http://%s", server.Addr)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return xerrors.Errorf("could not start a server: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down the server")
			if err := server.Shutdown(ctx); err != nil {
				return xerrors.Errorf("could not stop the server: %w", err)
			}
			return nil
		}
	})
	eg.Go(func() error {
		if err := execKubectl(ctx, portForwardArgs); err != nil {
			return xerrors.Errorf("could not run a port forwarder: %w", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("error while running the reverse proxy: %w", err)
	}
	return nil
}

func run(osArgs []string) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)
	go func() {
		<-signals
		cancel()
	}()

	var exitCode int
	f := genericclioptions.NewConfigFlags()
	rootCmd := cobra.Command{
		Use:     "kubectl oidc-port-forward TYPE/NAME [options] LOCAL_PORT:REMOTE_PORT",
		Short:   "Forward one or more local ports to a pod",
		Example: `  kubectl -n kube-system oidc-port-forward svc/kubernetes-dashboard 8443:443`,
		Args:    cobra.MinimumNArgs(2),
		Run: func(*cobra.Command, []string) {
			if err := runReverseProxyServer(ctx, f, osArgs[1:]); err != nil {
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

func main() {
	os.Exit(run(os.Args))
}
