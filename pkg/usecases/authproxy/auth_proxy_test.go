package authproxy

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/int128/kauthproxy/pkg/adaptors/env/mock_env"
	"github.com/int128/kauthproxy/pkg/adaptors/logger/mock_logger"
	"github.com/int128/kauthproxy/pkg/adaptors/portforwarder"
	"github.com/int128/kauthproxy/pkg/adaptors/portforwarder/mock_portforwarder"
	"github.com/int128/kauthproxy/pkg/adaptors/resolver/mock_resolver"
	"github.com/int128/kauthproxy/pkg/adaptors/reverseproxy"
	"github.com/int128/kauthproxy/pkg/adaptors/reverseproxy/mock_reverseproxy"
	"github.com/int128/kauthproxy/pkg/adaptors/transport"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var notNil = gomock.Not(gomock.Nil())

var restConfig rest.Config
var authProxyTransport http.Transport

func newTransport(t *testing.T) transport.NewFunc {
	return func(got *rest.Config) (http.RoundTripper, error) {
		if got != &restConfig {
			t.Errorf("rest.Config mismatch, got %+v", got)
		}
		return &authProxyTransport, nil
	}
}

func TestAuthProxy_Do(t *testing.T) {
	const containerPort = 18888
	const transitPort = 28888
	const podURL = "/api/v1/namespaces/kube-system/pods/kubernetes-dashboard-xxxxxxxx-xxxxxxxx"
	pod := &v1.Pod{
		ObjectMeta: v1meta.ObjectMeta{
			SelfLink: podURL,
		},
	}

	t.Run("ToPod", func(t *testing.T) {
		type mocks struct {
			resolverFactory *mock_resolver.MockFactoryInterface
			env             *mock_env.MockInterface
		}
		newMocks := func(ctrl *gomock.Controller) mocks {
			m := mocks{
				resolverFactory: mock_resolver.NewMockFactoryInterface(ctrl),
				env:             mock_env.NewMockInterface(ctrl),
			}
			m.env.EXPECT().
				AllocateLocalPort().
				Return(transitPort, nil)
			mockResolver := mock_resolver.NewMockInterface(ctrl)
			mockResolver.EXPECT().
				FindPodByName("NAMESPACE", "podname").
				Return(pod, containerPort, nil)
			m.resolverFactory.EXPECT().
				New(&restConfig).
				Return(mockResolver, nil)
			return m
		}

		t.Run("Success", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			portForwarder := mock_portforwarder.NewMockInterface(ctrl)
			portForwarder.EXPECT().
				Run(portforwarder.Option{
					Config:              &restConfig,
					SourcePort:          transitPort,
					TargetPodURL:        podURL,
					TargetContainerPort: containerPort,
				}, notNil, notNil).
				DoAndReturn(func(o portforwarder.Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
					time.Sleep(100 * time.Millisecond)
					close(readyChan)
					<-stopChan
					return nil
				})
			reverseProxy := mock_reverseproxy.NewMockInterface(ctrl)
			reverseProxy.EXPECT().
				Run(reverseproxy.Option{
					Transport:             &authProxyTransport,
					BindAddressCandidates: []string{"127.0.0.1:8000"},
					TargetScheme:          "https",
					TargetHost:            "localhost",
					TargetPort:            transitPort,
				}, notNil).
				DoAndReturn(func(o reverseproxy.Option, readyChan chan<- reverseproxy.Instance) error {
					time.Sleep(100 * time.Millisecond)
					i := mock_reverseproxy.NewMockInstance(ctrl)
					i.EXPECT().
						URL().
						Return(&url.URL{Scheme: "http", Host: "localhost:8000"})
					i.EXPECT().
						Shutdown(notNil).
						Return(nil)
					readyChan <- i
					return nil
				})
			m := newMocks(ctrl)
			m.env.EXPECT().
				OpenBrowser("http://localhost:8000")
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !xerrors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err wants context.DeadlineExceeded but was %+v", err)
			}
		})

		t.Run("PortForwarderError", func(t *testing.T) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			portForwarderError := xerrors.New("could not connect to pod")
			portForwarder := mock_portforwarder.NewMockInterface(ctrl)
			portForwarder.EXPECT().
				Run(portforwarder.Option{
					Config:              &restConfig,
					SourcePort:          transitPort,
					TargetPodURL:        podURL,
					TargetContainerPort: containerPort,
				}, notNil, notNil).
				DoAndReturn(func(o portforwarder.Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
					return portForwarderError
				})
			reverseProxy := mock_reverseproxy.NewMockInterface(ctrl)
			m := newMocks(ctrl)
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !xerrors.Is(err, portForwarderError) {
				t.Errorf("err wants the port forwarder error but was %+v", err)
			}
		})

		t.Run("ReverseProxyError", func(t *testing.T) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			portForwarder := mock_portforwarder.NewMockInterface(ctrl)
			portForwarder.EXPECT().
				Run(portforwarder.Option{
					Config:              &restConfig,
					SourcePort:          transitPort,
					TargetPodURL:        podURL,
					TargetContainerPort: containerPort,
				}, notNil, notNil).
				DoAndReturn(func(o portforwarder.Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
					time.Sleep(100 * time.Millisecond)
					close(readyChan)
					<-stopChan
					return nil
				})
			reverseProxyError := xerrors.New("could not listen")
			reverseProxy := mock_reverseproxy.NewMockInterface(ctrl)
			reverseProxy.EXPECT().
				Run(reverseproxy.Option{
					Transport:             &authProxyTransport,
					BindAddressCandidates: []string{"127.0.0.1:8000"},
					TargetScheme:          "https",
					TargetHost:            "localhost",
					TargetPort:            transitPort,
				}, notNil).
				DoAndReturn(func(o reverseproxy.Option, readyChan chan<- reverseproxy.Instance) error {
					return reverseProxyError
				})
			m := newMocks(ctrl)
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !xerrors.Is(err, reverseProxyError) {
				t.Errorf("err wants the port forwarder error but was %+v", err)
			}
		})

		t.Run("PortForwarderConnectionLost", func(t *testing.T) {
			// 0ms:   starting
			// 100ms: the port forwarder is ready
			// 200ms: the reverse proxy is ready
			// 400ms: lost connection
			// 900ms: retrying (after the backoff 500ms)
			// 1000ms: the port forwarder is ready
			// 1100ms: the reverse proxy is ready
			// 1200ms: cancel the context
			ctx, cancel := context.WithTimeout(context.TODO(), 1200*time.Millisecond)
			defer cancel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			portForwarder := mock_portforwarder.NewMockInterface(ctrl)
			portForwarder.EXPECT().
				Run(portforwarder.Option{
					Config:              &restConfig,
					SourcePort:          transitPort,
					TargetPodURL:        podURL,
					TargetContainerPort: containerPort,
				}, notNil, notNil).
				DoAndReturn(func(o portforwarder.Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
					time.Sleep(100 * time.Millisecond)
					close(readyChan)
					time.Sleep(300 * time.Millisecond)
					return nil // lost connection
				}).
				Times(2)
			reverseProxy := mock_reverseproxy.NewMockInterface(ctrl)
			reverseProxy.EXPECT().
				Run(reverseproxy.Option{
					Transport:             &authProxyTransport,
					BindAddressCandidates: []string{"127.0.0.1:8000"},
					TargetScheme:          "https",
					TargetHost:            "localhost",
					TargetPort:            transitPort,
				}, notNil).
				DoAndReturn(func(o reverseproxy.Option, readyChan chan<- reverseproxy.Instance) error {
					time.Sleep(100 * time.Millisecond)
					i := mock_reverseproxy.NewMockInstance(ctrl)
					i.EXPECT().
						URL().
						Return(&url.URL{Scheme: "http", Host: "localhost:8000"})
					i.EXPECT().
						Shutdown(notNil).
						Return(nil)
					readyChan <- i
					return nil
				}).
				Times(2)
			m := newMocks(ctrl)
			m.env.EXPECT().
				OpenBrowser("http://localhost:8000")
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !xerrors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err wants context.DeadlineExceeded but was %+v", err)
			}
		})
	})

	t.Run("ToService", func(t *testing.T) {
		type mocks struct {
			resolverFactory *mock_resolver.MockFactoryInterface
			env             *mock_env.MockInterface
		}
		newMocks := func(ctrl *gomock.Controller) mocks {
			m := mocks{
				resolverFactory: mock_resolver.NewMockFactoryInterface(ctrl),
				env:             mock_env.NewMockInterface(ctrl),
			}
			m.env.EXPECT().
				AllocateLocalPort().
				Return(transitPort, nil)
			mockResolver := mock_resolver.NewMockInterface(ctrl)
			mockResolver.EXPECT().
				FindPodByServiceName("NAMESPACE", "servicename").
				Return(pod, containerPort, nil)
			m.resolverFactory.EXPECT().
				New(&restConfig).
				Return(mockResolver, nil)
			return m
		}

		t.Run("Success", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
			defer cancel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			portForwarder := mock_portforwarder.NewMockInterface(ctrl)
			portForwarder.EXPECT().
				Run(portforwarder.Option{
					Config:              &restConfig,
					SourcePort:          transitPort,
					TargetPodURL:        podURL,
					TargetContainerPort: containerPort,
				}, notNil, notNil).
				DoAndReturn(func(o portforwarder.Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
					time.Sleep(100 * time.Millisecond)
					close(readyChan)
					<-stopChan
					return nil
				})
			reverseProxyInstance := mock_reverseproxy.NewMockInstance(ctrl)
			reverseProxyInstance.EXPECT().
				URL().
				Return(&url.URL{Scheme: "http", Host: "localhost:8000"})
			reverseProxyInstance.EXPECT().
				Shutdown(notNil).
				Return(nil)
			reverseProxy := mock_reverseproxy.NewMockInterface(ctrl)
			reverseProxy.EXPECT().
				Run(reverseproxy.Option{
					Transport:             &authProxyTransport,
					BindAddressCandidates: []string{"127.0.0.1:8000"},
					TargetScheme:          "https",
					TargetHost:            "localhost",
					TargetPort:            transitPort,
				}, notNil).
				DoAndReturn(func(o reverseproxy.Option, readyChan chan<- reverseproxy.Instance) error {
					time.Sleep(100 * time.Millisecond)
					readyChan <- reverseProxyInstance
					return nil
				})
			m := newMocks(ctrl)
			m.env.EXPECT().
				OpenBrowser("http://localhost:8000")
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://servicename.svc"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !xerrors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err wants context.DeadlineExceeded but was %+v", err)
			}
		})
	})
}

func parseURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		t.Errorf("could not parse URL: %s", err)
	}
	return u
}
