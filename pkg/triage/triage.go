package triage

// used to determine if the err from kube api is only because that type of resource is not in the targeted namespace
const KUBE_RESOURCE_NOT_FOUND string = "the server could not find the requested resource"

type Triage struct {
	ResourceType string   `yaml:"Resource"`
	AnomalyType  string   `yaml:"AnomalyType"`
	Anomalies    []string `yaml:"Anomalies"`
}

func NewTriage(resourceType string, anomalyType string, anomalies []string) *Triage {
	return &Triage{
		ResourceType: resourceType,
		AnomalyType:  anomalyType,
		Anomalies:    anomalies,
	}
}
