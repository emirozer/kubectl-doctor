package triage

import (
	"github.com/cheggaaa/pb/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

const componentHealthy = "True"

// TriageComponents gets a coreclient and checks if core components are in healthy state
// such as etcd cluster members, scheduler, controller-manager
func TriageComponents(coreClient coreclient.CoreV1Interface) (*Triage, error) {

	components, err := coreClient.ComponentStatuses().List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	bar := pb.StartNew(len(components.Items))
	listOfTriages := make([]string, 0)
	for _, i := range components.Items {
		bar.Increment()
		time.Sleep(time.Millisecond * 2)
		for _, y := range i.Conditions {
			if y.Status != componentHealthy {
				listOfTriages = append(listOfTriages, i.GetName())
			}
		}
	}
	bar.Finish()
	return NewTriage("ComponentStatuses", "Found unhealthy components!", listOfTriages), nil
}
