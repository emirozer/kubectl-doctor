package triage

import (
	"github.com/cheggaaa/pb/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

const pvcLostPhase = "Lost"

// TriagePVC gets a coreclient and checks if there are any pvcs that are in lost state
func TriagePVC(coreClient coreclient.CoreV1Interface) (*Triage, error) {
	listOfTriages := make([]string, 0)
	pvcs, err := coreClient.PersistentVolumeClaims("").List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	bar := pb.StartNew(len(pvcs.Items))
	for _, i := range pvcs.Items {
		bar.Increment()
		time.Sleep(time.Millisecond * 2)
		if i.Status.Phase == pvcLostPhase {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	bar.Finish()
	return NewTriage("PVC", "Found PVC in Lost State!", listOfTriages), nil
}
