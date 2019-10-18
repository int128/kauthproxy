// Package resolver provides resolving a pod and container port.
package resolver

import (
	"fmt"
	"strings"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/adaptors/logger"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var Set = wire.NewSet(
	wire.Struct(new(Factory), "*"),
	wire.Bind(new(FactoryInterface), new(*Factory)),
)

//go:generate mockgen -destination mock_resolver/mock_resolver.go github.com/int128/kauthproxy/pkg/adaptors/resolver FactoryInterface,Interface

type FactoryInterface interface {
	New(config *rest.Config) (Interface, error)
}

// Factory creates a Resolver.
type Factory struct {
	Logger logger.Interface
}

// New returns a Resolver.
func (f *Factory) New(config *rest.Config) (Interface, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, xerrors.Errorf("could not create a client: %w", err)
	}
	return &Resolver{
		Logger: f.Logger,
		CoreV1: clientset.CoreV1(),
	}, nil
}

type Interface interface {
	FindPodByServiceName(namespace, serviceName string) (*v1.Pod, int, error)
	FindPodByName(namespace, podName string) (*v1.Pod, int, error)
}

// Resolver provides resolving a pod and container port.
type Resolver struct {
	Logger logger.Interface
	CoreV1 corev1.CoreV1Interface
}

// FindPodByServiceName returns a pod and container port associated with the service.
func (r *Resolver) FindPodByServiceName(namespace, serviceName string) (*v1.Pod, int, error) {
	r.Logger.V(1).Infof("finding service %s in namespace %s", serviceName, namespace)
	service, err := r.CoreV1.Services(namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find the service: %w", err)
	}
	var selectors []string
	for k, v := range service.Spec.Selector {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(selectors, ",")
	r.Logger.V(1).Infof("finding pods by selector %s", selectors)
	pods, err := r.CoreV1.Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find pods by selector %s: %w", selector, err)
	}
	r.Logger.V(1).Infof("found %d pod(s)", len(pods.Items))
	if len(pods.Items) == 0 {
		return nil, 0, xerrors.Errorf("no pod matched to selector %s", selector)
	}
	pod := &pods.Items[0]
	r.Logger.V(1).Infof("first matched pod %s", pod.Name)
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			r.Logger.V(1).Infof("found container port %d in container %s of pod %s",
				port.ContainerPort, container.Name, pod.Name)
			return pod, int(port.ContainerPort), nil
		}
	}
	return nil, 0, xerrors.Errorf("no container port in pod %s", pod.Name)
}

// FindPodByName finds a pod and container port by name.
func (r *Resolver) FindPodByName(namespace, podName string) (*v1.Pod, int, error) {
	r.Logger.V(1).Infof("finding pod %s in namespace %s", podName, namespace)
	pod, err := r.CoreV1.Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return nil, 0, xerrors.Errorf("could not find the pod: %w", err)
	}
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			r.Logger.V(1).Infof("found container port %d in container %s of pod %s",
				port.ContainerPort, container.Name, pod.Name)
			return pod, int(port.ContainerPort), nil
		}
	}
	return nil, 0, xerrors.Errorf("no container port in pod %s", pod.Name)
}
