package triage

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

const pvAvailable = "Available"

// TriagePV gets a coreclient and checks if there are any pvs that are Available and Unclaimed
func TriagePV(coreClient coreclient.CoreV1Interface) (*Triage, error) {
	listOfTriages := make([]string, 0)
	pvs, err := coreClient.PersistentVolumes().List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, i := range pvs.Items {
		if i.Status.Phase == pvAvailable {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	return NewTriage("PV", "Found PV in Available & Unclaimed State!", listOfTriages), nil
}
