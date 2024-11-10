package authproxy

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/int128/kauthproxy/internal/logger/mock_logger"
	"github.com/int128/kauthproxy/internal/mocks/mock_browser"
	"github.com/int128/kauthproxy/internal/mocks/mock_env"
	"github.com/int128/kauthproxy/internal/mocks/mock_portforwarder"
	"github.com/int128/kauthproxy/internal/mocks/mock_resolver"
	"github.com/int128/kauthproxy/internal/mocks/mock_reverseproxy"
	"github.com/int128/kauthproxy/internal/portforwarder"
	"github.com/int128/kauthproxy/internal/reverseproxy"
	"github.com/int128/kauthproxy/internal/transport"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubernetes-dashboard-12345678-12345678",
			Namespace: "kubernetes-dashboard",
		},
	}

	t.Run("ToPod", func(t *testing.T) {
		type mocks struct {
			resolverFactory *mock_resolver.MockFactoryInterface
			env             *mock_env.MockInterface
			browser         *mock_browser.MockInterface
		}
		newMocks := func(ctrl *gomock.Controller) mocks {
			m := mocks{
				resolverFactory: mock_resolver.NewMockFactoryInterface(ctrl),
				env:             mock_env.NewMockInterface(ctrl),
				browser:         mock_browser.NewMockInterface(ctrl),
			}
			m.env.EXPECT().
				AllocateLocalPort().
				Return(transitPort, nil)
			mockResolver := mock_resolver.NewMockInterface(ctrl)
			mockResolver.EXPECT().
				FindPodByName(gomock.Any(), "NAMESPACE", "podname").
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
					TargetNamespace:     "kubernetes-dashboard",
					TargetPodName:       "kubernetes-dashboard-12345678-12345678",
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
			m.browser.EXPECT().Open("http://localhost:8000")
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Browser:         m.browser,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err wants context.DeadlineExceeded but was %+v", err)
			}
		})

		t.Run("PortForwarderError", func(t *testing.T) {
			ctx := context.TODO()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			portForwarderError := errors.New("could not connect to pod")
			portForwarder := mock_portforwarder.NewMockInterface(ctrl)
			portForwarder.EXPECT().
				Run(portforwarder.Option{
					Config:              &restConfig,
					SourcePort:          transitPort,
					TargetNamespace:     "kubernetes-dashboard",
					TargetPodName:       "kubernetes-dashboard-12345678-12345678",
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
				Browser:         m.browser,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !errors.Is(err, portForwarderError) {
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
					TargetNamespace:     "kubernetes-dashboard",
					TargetPodName:       "kubernetes-dashboard-12345678-12345678",
					TargetContainerPort: containerPort,
				}, notNil, notNil).
				DoAndReturn(func(o portforwarder.Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
					time.Sleep(100 * time.Millisecond)
					close(readyChan)
					<-stopChan
					return nil
				})
			reverseProxyError := errors.New("could not listen")
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
				Browser:         m.browser,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !errors.Is(err, reverseProxyError) {
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
					TargetNamespace:     "kubernetes-dashboard",
					TargetPodName:       "kubernetes-dashboard-12345678-12345678",
					TargetContainerPort: containerPort,
				}, notNil, notNil).
				DoAndReturn(func(o portforwarder.Option, readyChan chan struct{}, stopChan <-chan struct{}) error {
					time.Sleep(100 * time.Millisecond)
					close(readyChan)
					time.Sleep(300 * time.Millisecond)
					return nil // lost connection
				}).
				MinTimes(2)
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
				MinTimes(2)
			m := newMocks(ctrl)
			m.browser.EXPECT().Open("http://localhost:8000")
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Browser:         m.browser,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://podname"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err wants context.DeadlineExceeded but was %+v", err)
			}
		})
	})

	t.Run("ToService", func(t *testing.T) {
		type mocks struct {
			resolverFactory *mock_resolver.MockFactoryInterface
			env             *mock_env.MockInterface
			browser         *mock_browser.MockInterface
		}
		newMocks := func(ctrl *gomock.Controller) mocks {
			m := mocks{
				resolverFactory: mock_resolver.NewMockFactoryInterface(ctrl),
				env:             mock_env.NewMockInterface(ctrl),
				browser:         mock_browser.NewMockInterface(ctrl),
			}
			m.env.EXPECT().
				AllocateLocalPort().
				Return(transitPort, nil)
			mockResolver := mock_resolver.NewMockInterface(ctrl)
			mockResolver.EXPECT().
				FindPodByServiceName(gomock.Any(), "NAMESPACE", "servicename").
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
					TargetNamespace:     "kubernetes-dashboard",
					TargetPodName:       "kubernetes-dashboard-12345678-12345678",
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
			m.browser.EXPECT().Open("http://localhost:8000")
			u := &AuthProxy{
				ReverseProxy:    reverseProxy,
				PortForwarder:   portForwarder,
				ResolverFactory: m.resolverFactory,
				NewTransport:    newTransport(t),
				Env:             m.env,
				Browser:         m.browser,
				Logger:          mock_logger.New(t),
			}
			o := Option{
				Config:                &restConfig,
				Namespace:             "NAMESPACE",
				TargetURL:             parseURL(t, "https://servicename.svc"),
				BindAddressCandidates: []string{"127.0.0.1:8000"},
			}
			err := u.Do(ctx, o)
			if !errors.Is(err, context.DeadlineExceeded) {
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
