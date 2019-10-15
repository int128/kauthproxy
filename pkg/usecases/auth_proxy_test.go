package usecases

import (
	"context"
	"golang.org/x/xerrors"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/int128/kauthproxy/pkg/adaptors/logger/mock_logger"
	"github.com/int128/kauthproxy/pkg/adaptors/network/mock_network"
	"github.com/int128/kauthproxy/pkg/adaptors/portforwarder"
	"github.com/int128/kauthproxy/pkg/adaptors/portforwarder/mock_portforwarder"
	"github.com/int128/kauthproxy/pkg/adaptors/resolver/mock_resolver"
	"github.com/int128/kauthproxy/pkg/adaptors/reverseproxy"
	"github.com/int128/kauthproxy/pkg/adaptors/reverseproxy/mock_reverseproxy"
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
			Run(notNil, reverseproxy.Options{
				Transport: authProxyTransport,
				Source: reverseproxy.Source{
					AddressCandidates: []string{"127.0.0.1:8000"},
				},
				Target: reverseproxy.Target{
					Scheme: "https",
					Host:   "localhost",
					Port:   transitPort,
				},
			}).
			Return(xerrors.Errorf("finally context canceled: %w", context.Canceled))
		portForwarder := mock_portforwarder.NewMockInterface(ctrl)
		portForwarder.EXPECT().
			Run(notNil, portforwarder.Options{
				Config: c,
				Source: portforwarder.Source{Port: transitPort},
				Target: portforwarder.Target{
					Pod:           pod,
					ContainerPort: containerPort,
				},
			}).
			Return(xerrors.Errorf("finally context canceled: %w", context.Canceled))
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
			Config:                c,
			Namespace:             "NAMESPACE",
			TargetURL:             parseURL(t, "https://podname"),
			BindAddressCandidates: []string{"127.0.0.1:8000"},
		}
		err := u.Do(context.TODO(), o)
		if !xerrors.Is(err, context.Canceled) {
			t.Errorf("err wants context.Canceled but was %+v", err)
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
			Run(notNil, reverseproxy.Options{
				Transport: authProxyTransport,
				Source: reverseproxy.Source{
					AddressCandidates: []string{"127.0.0.1:8000"},
				},
				Target: reverseproxy.Target{
					Scheme: "https",
					Host:   "localhost",
					Port:   transitPort,
				},
			}).
			Return(xerrors.Errorf("finally context canceled: %w", context.Canceled))
		portForwarder := mock_portforwarder.NewMockInterface(ctrl)
		portForwarder.EXPECT().
			Run(notNil, portforwarder.Options{
				Config: c,
				Source: portforwarder.Source{Port: transitPort},
				Target: portforwarder.Target{
					Pod:           pod,
					ContainerPort: containerPort,
				},
			}).
			Return(xerrors.Errorf("finally context canceled: %w", context.Canceled))
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
			Config:                c,
			Namespace:             "NAMESPACE",
			TargetURL:             parseURL(t, "https://servicename.svc"),
			BindAddressCandidates: []string{"127.0.0.1:8000"},
		}
		err := u.Do(context.TODO(), o)
		if !xerrors.Is(err, context.Canceled) {
			t.Errorf("err wants context.Canceled but was %+v", err)
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
