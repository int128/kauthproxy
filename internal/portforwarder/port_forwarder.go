// Package portforwarder provides port forwarding between local and Kubernetes.
package portforwarder

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/google/wire"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var Set = wire.NewSet(
	wire.Struct(new(PortForwarder), "*"),
	wire.Bind(new(Interface), new(*PortForwarder)),
)

//go:generate go run -mod=mod go.uber.org/mock/mockgen -destination mock_portforwarder/mock_portforwarder.go github.com/int128/kauthproxy/internal/portforwarder Interface

// Option represents an option of PortForwarder.
type Option struct {
	Config              *rest.Config
	SourcePort          int
	TargetNamespace     string
	TargetPodName       string
	TargetContainerPort int
}

type Interface interface {
	Run(o Option, readyChan chan struct{}, stopChan <-chan struct{}) error
}

type PortForwarder struct {
}

// Run executes a port forwarder.
//
// It returns nil if stopChan has been closed or connection has lost.
// It returns an error if it could not connect to the pod.
//
// It will close the readyChan when the port forwarder is ready.
// Caller can stop the port forwarder by closing the stopChan.
func (pf *PortForwarder) Run(o Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
	pfURL, err := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", o.Config.Host, o.TargetNamespace, o.TargetPodName))
	if err != nil {
		return fmt.Errorf("could not build URL for portforward: %w", err)
	}
	rt, upgrader, err := spdy.RoundTripperFor(o.Config)
	if err != nil {
		return fmt.Errorf("could not create a round tripper: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: rt}, http.MethodPost, pfURL)
	portPair := fmt.Sprintf("%d:%d", o.SourcePort, o.TargetContainerPort)
	forwarder, err := portforward.NewOnAddresses(dialer, []string{"127.0.0.1"}, []string{portPair}, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("could not create a port forwarder: %w", err)
	}
	if err := forwarder.ForwardPorts(); err != nil {
		return fmt.Errorf("could not run the port forwarder at %s: %w", portPair, err)
	}
	return nil
}
