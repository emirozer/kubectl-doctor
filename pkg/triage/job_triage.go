package triage

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LeftoverJobs gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover cronjobs that were inactive for more than a month
func LeftoverJobs(ctx context.Context, kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)

	jobs, err := kubeCli.BatchV1().CronJobs(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		if err.Error() != KUBE_RESOURCE_NOT_FOUND {
			return nil, err
		}
	}

	currentTime := time.Now()
	for _, i := range jobs.Items {
		if i.Status.LastScheduleTime != nil {
			if currentTime.Sub(i.Status.LastScheduleTime.Local()).Hours()/24 > 30 {
				listOfTriages = append(listOfTriages, i.GetName())
			}
		}

	}
	return NewTriage("CronJobs", "Found leftover cronjobs in namespace: "+namespace, listOfTriages), nil
}
