package triage

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LeftoverIngresses gets a kubernetes.Clientset and a specific namespace string
// then proceeds to search if there are leftover ingresses
func LeftoverIngresses(kubeCli *kubernetes.Clientset, namespace string) (*Triage, error) {
	listOfTriages := make([]string, 0)

	ingresses, err := kubeCli.NetworkingV1beta1().Ingresses(namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, i := range ingresses.Items {
		fmt.Println("-------------")
		fmt.Println(i.Status)
		fmt.Println("-------------")
	}
	return NewTriage("Ingress", "Found leftover ingresses in namespace: "+namespace, listOfTriages), nil
}
