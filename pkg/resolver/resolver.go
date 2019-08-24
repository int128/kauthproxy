// Package resolver provides resolving a pod and container port.
package resolver

import (
	"fmt"
	"strings"

	"github.com/google/wire"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var Set = wire.NewSet(
	wire.Struct(new(Factory)),
	wire.Bind(new(FactoryInterface), new(*Factory)),
)

//go:generate mockgen -destination mock_resolver/mock_resolver.go github.com/int128/kauthproxy/pkg/resolver FactoryInterface,Interface

type FactoryInterface interface {
	New(config *rest.Config) (Interface, error)
}

// Factory creates a Resolver.
type Factory struct{}

// New returns a Resolver.
func (*Factory) New(config *rest.Config) (Interface, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, xerrors.Errorf("could not create a client: %w", err)
	}
	return &Resolver{CoreV1: clientset.CoreV1()}, nil
}

type Interface interface {
	FindByServiceName(namespace, serviceName string) (*v1.Pod, int, error)
	FindByPodName(namespace, podName string) (*v1.Pod, int, error)
}

// Resolver provides resolving a pod and container port.
type Resolver struct {
	CoreV1 corev1.CoreV1Interface
}

// FindByServiceName returns a pod and container port associated with the service.
func (r *Resolver) FindByServiceName(namespace, serviceName string) (*v1.Pod, int, error) {
	service, err := r.CoreV1.Services(namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find the service: %w", err)
	}
	var selectors []string
	for k, v := range service.Spec.Selector {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(selectors, ",")
	pods, err := r.CoreV1.Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find pods by selector %s: %w", selector, err)
	}
	if len(pods.Items) == 0 {
		return nil, 0, xerrors.Errorf("no pod matched to selector %s", selector)
	}
	pod := &pods.Items[0]
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			return pod, int(port.ContainerPort), nil
		}
	}
	return nil, 0, xerrors.Errorf("no container port in the pod %s", pod.Name)
}

// FindByPodName finds a pod and container port by name.
func (r *Resolver) FindByPodName(namespace, podName string) (*v1.Pod, int, error) {
	pod, err := r.CoreV1.Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find the pod: %w", err)
	}
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			return pod, int(port.ContainerPort), nil
		}
	}
	return nil, 0, xerrors.Errorf("no container port in the pod %s", pod.Name)
}
