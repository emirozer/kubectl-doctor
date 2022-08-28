package triage

import (
	"context"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StandalonePods gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover deployments
// the criteria is that a pod has no ownership (Deployment/Statefulset)
func TerminatingNamespaces(ctx context.Context, kubeCli *kubernetes.Clientset) (*Triage, error) {
	listOfTriages := make([]string, 0)
	namespaces, err := kubeCli.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), KUBE_RESOURCE_NOT_FOUND) {
			return nil, err
		}
	}
	for _, i := range namespaces.Items {
		if i.Status.Phase == "Terminating" {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("Namespaces", "Found Terminating Namespaces: ", listOfTriages), nil
}
