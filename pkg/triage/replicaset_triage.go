package triage

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// OrphanedReplicaSet gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are orphan replicasets
// the criteria is that the desired number of replicas are bigger than 0 but the available replicas are 0
func OrphanedReplicaSet(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)
	rs, err := kubeCli.AppsV1().ReplicaSets(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if err.Error() != KUBE_RESOURCE_NOT_FOUND {
			return nil, err
		}
	}

	for _, i := range rs.Items {
		if i.Status.Replicas > 0 && i.Status.AvailableReplicas == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("ReplicaSets", "Found orphan replicasets in namespace: "+namespace, listOfTriages), nil
}

// LeftOverReplicaSet gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are left over replicasets
// the criteria is that both the desired number of replicas and the available # of replicas are 0
func LeftOverReplicaSet(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)
	rs, err := kubeCli.AppsV1().ReplicaSets(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if err.Error() != KUBE_RESOURCE_NOT_FOUND {
			return nil, err
		}
	}

	for _, i := range rs.Items {
		if i.Status.Replicas == 0 && i.Status.AvailableReplicas == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("ReplicaSets", "Found leftover replicasets in namespace: "+namespace, listOfTriages), nil
}
