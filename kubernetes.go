package pprofserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/events"
	apiv1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// KubernetesRegistry is a registry which discovers PODs running
// on a Kubernetes cluster.
//
// TODO: give the ability to configure multiple Kubernetes clusters.
type KubernetesRegistry struct {
	Namespace string

	client kubernetes.Interface
	store  cache.Store
}

func NewKubernetesRegistry(client *kubernetes.Clientset) *KubernetesRegistry {
	return &KubernetesRegistry{
		client: client,
	}
}

// Name implements the Registry interface.
func (k *KubernetesRegistry) String() string {
	return "kubernetes"
}

// Init initialize the watcher and store configuration for the registry.
func (k *KubernetesRegistry) Init(ctx context.Context) {
	p := k.client.CoreV1().Pods(k.Namespace)

	listWatch := &cache.ListWatch{
		ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
			return p.List(options)
		},
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			return p.Watch(options)
		},
	}

	queue := workqueue.New()

	informer := cache.NewSharedInformer(listWatch, &apiv1.Pod{}, 10*time.Second)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			k.handleObj(queue, obj)
		},
		DeleteFunc: func(obj interface{}) {
			k.handleObj(queue, obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			k.handleObj(queue, obj)
		},
	})

	go informer.Run(ctx.Done())

	k.store = informer.GetStore()
}

func (k *KubernetesRegistry) handleObj(q *workqueue.Type, o interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(o)
	if err != nil {
		events.Log("failed to handle object: %{error}s", err)
		return
	}

	q.Add(key)
}

func toPod(o interface{}) (*apiv1.Pod, error) {
	pod, ok := o.(*apiv1.Pod)
	if ok {
		return pod, nil
	}

	return nil, fmt.Errorf("received unexpected object: %v", o)
}

func (k *KubernetesRegistry) ListServices(ctx context.Context) ([]string, error) {

	podnames, err := k.client.CoreV1().Pods("").List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	list := make([]string, 0, len(podnames.Items))
	for _, pod := range podnames.Items {
		list = append(list, pod.Name)
	}

	return list, nil
}

// LookupService implements the Registry interface. The returned Service will contain
// one Host entry per POD IP+container exposed port.
func (k *KubernetesRegistry) LookupService(ctx context.Context, name string) (Service, error) {
	svc := Service{
		Name: "kubernetes",
	}

	hosts := []Host{}
	for _, obj := range k.store.List() {
		pod, err := toPod(obj)
		if err != nil {
			events.Log("failed to convert data to pod: %{error}s", err)
			continue
		}
		// filtering pods based on podname, even if they are diff namepsaces for now, since the route for namespaces isnt made yet
		if pod.Name == name {
			for _, container := range pod.Spec.Containers {
				// adding container name to display
				tags := []string{pod.Name + "-" + container.Name}

				for _, port := range container.Ports {
					if port.Name == "http" {
						hosts = append(hosts, Host{
							Addr: &net.TCPAddr{
								IP:   net.ParseIP(pod.Status.PodIP),
								Port: int(port.ContainerPort),
							},
							Tags: append(tags, port.Name), // port name must be specified in the pod spec as http
						})
					}
				}
			}
		}
	}

	svc.Hosts = hosts

	return svc, nil
}
