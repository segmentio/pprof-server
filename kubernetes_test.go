package pprofserver

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func makePod() *v1.Pod {
	return &v1.Pod{
		Spec: v1.PodSpec{
			NodeName: "instance0",
			Containers: []v1.Container{
				{
					Name: "container0",
					Ports: []v1.ContainerPort{
						{
							Name:          "http",
							Protocol:      v1.ProtocolTCP,
							ContainerPort: int32(3000),
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			PodIP:  "1.2.3.4",
			HostIP: "1.2.3.5",
		},
	}
}

func kubernetesRegistryWith(objects ...runtime.Object) *KubernetesRegistry {
	return &KubernetesRegistry{
		client: fake.NewSimpleClientset(objects...),
	}
}

func TestKubernetesRegistry(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		scenario string
		objects  []runtime.Object
	}{
		{"single valid pod", []runtime.Object{makePod()}},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			registry := kubernetesRegistryWith(test.objects...)
			registry.Init(ctx)
			_, err := registry.LookupService(ctx, "")
			if err != nil {
				t.Error(err)
			}
		})
	}
}
