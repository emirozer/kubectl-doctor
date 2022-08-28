package triage

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

const nodeTargetReason = "KubeletReady"

// TriageNodes gets a coreclient for k8s and checks if there are any nodes in the cluster
// that are not in Ready state(unoperational nodes)
func TriageNodes(ctx context.Context, coreClient coreclient.CoreV1Interface) (*Triage, error) {
	listOfTriages := make([]string, 0)
	nodes, err := coreClient.Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		if err.Error() != KUBE_RESOURCE_NOT_FOUND {
			return nil, err
		}
	}

	for _, i := range nodes.Items {
		for _, y := range i.Status.Conditions {
			if y.Reason == nodeTargetReason {
				if y.Status != "True" {
					listOfTriages = append(listOfTriages, i.GetName())
					break
				}
			}
		}
	}
	return NewTriage("Nodes", "Found node(s) that are not in Ready state!", listOfTriages), nil
}
