package triage

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
