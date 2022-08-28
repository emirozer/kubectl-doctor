package triage

import (
	"context"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const podTargetReason = "Ready"

// StandalonePods gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover deployments
// the criteria is that a pod has no ownership (Deployment/Statefulset)
func StandalonePods(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)
	pods, err := kubeCli.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), KUBE_RESOURCE_NOT_FOUND) {
			return nil, err
		}
	}
	for _, i := range pods.Items {
		if len(i.GetOwnerReferences()) == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("Pods", "Found standalone pods in namespace: "+namespace, listOfTriages), nil
}

// UnreadyPods gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover deployments
// the criteria is that a pod has no ownership (Deployment/Statefulset)
func UnreadyPods(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)
	pods, err := kubeCli.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), KUBE_RESOURCE_NOT_FOUND) {
			return nil, err
		}
	}

	for _, i := range pods.Items {
		for _, y := range i.Status.Conditions {
			if y.Reason == podTargetReason {
				if y.Status != "True" {
					listOfTriages = append(listOfTriages, i.GetName())
					break
				}
			}
		}
	}
	return NewTriage("Pods", "Found Unready pods in namespace: "+namespace, listOfTriages), nil
}
