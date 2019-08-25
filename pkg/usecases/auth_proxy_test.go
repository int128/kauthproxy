package usecases

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/int128/kauthproxy/pkg/logger/mock_logger"
	"github.com/int128/kauthproxy/pkg/network/mock_network"
	"github.com/int128/kauthproxy/pkg/portforwarder"
	"github.com/int128/kauthproxy/pkg/portforwarder/mock_portforwarder"
	"github.com/int128/kauthproxy/pkg/resolver/mock_resolver"
	"github.com/int128/kauthproxy/pkg/reverseproxy"
	"github.com/int128/kauthproxy/pkg/reverseproxy/mock_reverseproxy"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

var notNil = gomock.Not(gomock.Nil())

func TestAuthProxy_Do(t *testing.T) {
	t.Run("ProxyToPod", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := &rest.Config{}
		pod := &v1.Pod{}
		authProxyTransport := &http.Transport{}
		const containerPort = 18888
		const transitPort = 28888

		reverseProxy := mock_reverseproxy.NewMockInterface(ctrl)
		reverseProxy.EXPECT().
			Start(notNil, notNil, reverseproxy.Options{
				Transport: authProxyTransport,
				Source:    reverseproxy.Source{Address: "localhost:8888"},
				Target: reverseproxy.Target{
					Scheme: "https",
					Host:   "localhost",
					Port:   transitPort,
				},
			})
		portForwarder := mock_portforwarder.NewMockInterface(ctrl)
		portForwarder.EXPECT().
			Start(notNil, notNil, portforwarder.Options{
				Config: c,
				Source: portforwarder.Source{Port: transitPort},
				Target: portforwarder.Target{
					Pod:           pod,
					ContainerPort: containerPort,
				},
			})
		mockResolver := mock_resolver.NewMockInterface(ctrl)
		mockResolver.EXPECT().
			FindByPodName("NAMESPACE", "podname").
			Return(pod, containerPort, nil)
		resolverFactory := mock_resolver.NewMockFactoryInterface(ctrl)
		resolverFactory.EXPECT().
			New(c).
			Return(mockResolver, nil)
		mockNetwork := mock_network.NewMockInterface(ctrl)
		mockNetwork.EXPECT().
			AllocateLocalPort().
			Return(transitPort, nil)
		mockNetwork.EXPECT().
			NewTransportWithToken(c).
			Return(authProxyTransport, nil)

		u := &AuthProxy{
			ReverseProxy:    reverseProxy,
			PortForwarder:   portForwarder,
			ResolverFactory: resolverFactory,
			Network:         mockNetwork,
			Logger:          mock_logger.New(t),
		}
		o := AuthProxyOptions{
			Config:    c,
			Namespace: "NAMESPACE",
			RemoteURL: parseURL(t, "https://podname"),
			LocalAddr: "localhost:8888",
		}
		if err := u.Do(context.Background(), o); err != nil {
			t.Errorf("err wants nil but was %+v", err)
		}
	})
	t.Run("ProxyToService", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := &rest.Config{}
		pod := &v1.Pod{}
		authProxyTransport := &http.Transport{}
		const containerPort = 19999
		const transitPort = 29999

		reverseProxy := mock_reverseproxy.NewMockInterface(ctrl)
		reverseProxy.EXPECT().
			Start(notNil, notNil, reverseproxy.Options{
				Transport: authProxyTransport,
				Source:    reverseproxy.Source{Address: "localhost:9999"},
				Target: reverseproxy.Target{
					Scheme: "https",
					Host:   "localhost",
					Port:   transitPort,
				},
			})
		portForwarder := mock_portforwarder.NewMockInterface(ctrl)
		portForwarder.EXPECT().
			Start(notNil, notNil, portforwarder.Options{
				Config: c,
				Source: portforwarder.Source{Port: transitPort},
				Target: portforwarder.Target{
					Pod:           pod,
					ContainerPort: containerPort,
				},
			})
		mockResolver := mock_resolver.NewMockInterface(ctrl)
		mockResolver.EXPECT().
			FindByServiceName("NAMESPACE", "servicename").
			Return(pod, containerPort, nil)
		resolverFactory := mock_resolver.NewMockFactoryInterface(ctrl)
		resolverFactory.EXPECT().
			New(c).
			Return(mockResolver, nil)
		mockNetwork := mock_network.NewMockInterface(ctrl)
		mockNetwork.EXPECT().
			AllocateLocalPort().
			Return(transitPort, nil)
		mockNetwork.EXPECT().
			NewTransportWithToken(c).
			Return(authProxyTransport, nil)

		u := &AuthProxy{
			ReverseProxy:    reverseProxy,
			PortForwarder:   portForwarder,
			ResolverFactory: resolverFactory,
			Network:         mockNetwork,
			Logger:          mock_logger.New(t),
		}
		o := AuthProxyOptions{
			Config:    c,
			Namespace: "NAMESPACE",
			RemoteURL: parseURL(t, "https://servicename.svc"),
			LocalAddr: "localhost:9999",
		}
		if err := u.Do(context.Background(), o); err != nil {
			t.Errorf("err wants nil but was %+v", err)
		}
	})
}

func parseURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		t.Errorf("could not parse URL: %s", err)
	}
	return u
}
