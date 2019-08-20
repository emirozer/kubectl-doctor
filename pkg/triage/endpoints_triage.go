package triage

import (
	"github.com/cheggaaa/pb/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

// TriageEndpoints gets a coreclient for k8s and scans through all endpoints to see if they are leftover/unused
func TriageEndpoints(coreClient coreclient.CoreV1Interface) (*Triage, error) {
	endpoints, err := coreClient.Endpoints("").List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	bar := pb.StartNew(len(endpoints.Items))
	listOfTriages := make([]string, 0)
	for _, i := range endpoints.Items {
		bar.Increment()
		time.Sleep(time.Millisecond * 2)
		if len(i.Subsets) == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	bar.Finish()
	return NewTriage("Endpoints", "Found orphaned endpoints!", listOfTriages), nil
}
