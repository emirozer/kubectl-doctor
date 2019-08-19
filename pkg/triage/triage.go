package triage

type Triage struct {
	ResourceType string
	AnomalyType  string
	Anomalies    []string
}

func NewTriage(resourceType string, anomalyType string, anomalies []string) *Triage {
	return &Triage{
		ResourceType: resourceType,
		AnomalyType:  anomalyType,
		Anomalies:    anomalies,
	}
}
