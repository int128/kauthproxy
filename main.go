package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

func runProxy(f *genericclioptions.ConfigFlags) error {
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
	if err := server.ListenAndServe(); err != nil {
		return xerrors.Errorf("could not start a server: %w", err)
	}
	return nil
}

func run(osArgs []string) int {
	var exitCode int
	f := genericclioptions.NewConfigFlags()
	rootCmd := cobra.Command{
		Use: filepath.Base(osArgs[0]),
		Run: func(*cobra.Command, []string) {
			if err := runProxy(f); err != nil {
				log.Printf("error: %s", err)
				exitCode = 1
			}
		},
	}
	f.AddFlags(rootCmd.Flags())

	rootCmd.SetArgs(osArgs[1:])
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return exitCode
}

func main() {
	os.Exit(run(os.Args))
}
