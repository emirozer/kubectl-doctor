package triage

import (
	"github.com/cheggaaa/pb/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

// TriageDeployments gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover deployments
// the criteria is that the desired number of replicas are bigger than 0 but the available replicas are 0
func TriageDeployments(kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)
	deployments, err := kubeCli.ExtensionsV1beta1().Deployments(namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	bar := pb.StartNew(len(deployments.Items))
	for _, i := range deployments.Items {
		bar.Increment()
		time.Sleep(time.Millisecond * 2)
		if i.Status.Replicas > 0 && i.Status.AvailableReplicas == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	bar.Finish()
	return NewTriage("Deployments", "Found leftover deployments!", listOfTriages), nil
}
