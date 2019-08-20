package triage

import (
	"github.com/cheggaaa/pb/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

const pvAvailable = "Available"

// TriagePV gets a coreclient and checks if there are any pvs that are Available and Unclaimed
func TriagePV(coreClient coreclient.CoreV1Interface) (*Triage, error) {
	listOfTriages := make([]string, 0)
	pvs, err := coreClient.PersistentVolumes().List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	bar := pb.StartNew(len(pvs.Items))
	for _, i := range pvs.Items {
		bar.Increment()
		time.Sleep(time.Millisecond * 2)
		if i.Status.Phase == pvAvailable {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	bar.Finish()
	return NewTriage("PV", "Found PV in Available & Unclaimed State!", listOfTriages), nil
}
