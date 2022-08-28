package triage

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

// TriageEndpoints gets a coreclient for k8s and scans through all endpoints to see if they are leftover/unused
func TriageEndpoints(ctx context.Context, coreClient coreclient.CoreV1Interface, namespace string) (*Triage, error) {
	endpoints, err := coreClient.Endpoints(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if err.Error() != KUBE_RESOURCE_NOT_FOUND {
			return nil, err
		}
	}

	listOfTriages := make([]string, 0)
	for _, i := range endpoints.Items {
		if len(i.Subsets) == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("Endpoints", "Found orphaned endpoints in namespace: "+namespace, listOfTriages), nil
}
