package triage

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LeftoverIngresses gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover ingresses
func LeftoverIngresses(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)

	ingresses, err := kubeCli.NetworkingV1().Ingresses(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if err.Error() != KUBE_RESOURCE_NOT_FOUND {
			return nil, err
		}
	}

	for _, i := range ingresses.Items {
		if i.Status.LoadBalancer.Size() <= 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}

	}
	return NewTriage("Ingress", "Found leftover ingresses in namespace: "+namespace, listOfTriages), nil
}
