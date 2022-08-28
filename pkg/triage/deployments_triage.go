package triage

import (
	"context"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// OrphanedDeployments gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover deployments
// the criteria is that the desired number of replicas are bigger than 0 but the available replicas are 0
func OrphanedDeployments(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)
	deployments, err := kubeCli.AppsV1().Deployments(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), KUBE_RESOURCE_NOT_FOUND) {
			return nil, err
		}
	}
	for _, i := range deployments.Items {
		if i.Status.Replicas > 0 && i.Status.AvailableReplicas == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("Deployments", "Found orphan deployments in namespace: "+namespace, listOfTriages), nil
}

// LeftOverDeployments gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover deployments
// the criteria is that both the desired number of replicas and the available # of replicas are 0
func LeftOverDeployments(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)
	deployments, err := kubeCli.AppsV1().Deployments(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), KUBE_RESOURCE_NOT_FOUND) {
			return nil, err
		}
	}

	for _, i := range deployments.Items {
		if i.Status.Replicas == 0 && i.Status.AvailableReplicas == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("Deployments", "Found leftover deployments in namespace: "+namespace, listOfTriages), nil
}
