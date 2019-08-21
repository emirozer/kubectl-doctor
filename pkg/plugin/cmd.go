package plugin

import (
	"github.com/emirozer/kubectl-doctor/pkg/client"
	"github.com/emirozer/kubectl-doctor/pkg/triage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
)

const (
	example = `
	# triage everything in a given target namespace
	kubectl doctor my_namespace
    # triage only deployments in a given target namespace
    kubectl doctor my_namespace --deployment-only
`
	longDesc = `
    kubectl-doctor plugin will scan the given namespace for any kind of anomalies and reports back to its user.
    example anomalies: 
        * deployments that are older than 30d with 0 available, 
        * deployments that do not have minimum availability,
        * kubernetes nodes cpu usage or memory usage too high. or too low to report scaledown possiblity 
`

	usageError = "expects 'doctor NAMESPACE' for doctor command"
)

var (
	clientset *kubernetes.Clientset
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	// Only log the info severity or above.
	log.SetLevel(log.InfoLevel)
	clientset = client.InitClient()
}

// DoctorOptions specify what the doctor is going to do
type DoctorOptions struct {
	FetchedNamespaces []string

	// Doctor options
	DeploymentOnly bool
	FullScan       bool
	Flags          *genericclioptions.ConfigFlags
	CoreClient     coreclient.CoreV1Interface
	RESTClient     *restclient.RESTClient
	KubeCli        *kubernetes.Clientset
	Args           []string
	Config         *restclient.Config
}

// NewDoctorOptions new doctor options initializer
func NewDoctorOptions() *DoctorOptions {
	return &DoctorOptions{
		Flags: genericclioptions.NewConfigFlags(true),
	}
}

// NewDoctorCmd returns a cobra command wrapping DoctorOptions
func NewDoctorCmd() *cobra.Command {

	opts := NewDoctorOptions()

	cmd := &cobra.Command{
		Use:     "doctor [-n NAMESPACE] -- COMMAND [args...]",
		Short:   "start triage for current targeted kubernetes cluster",
		Long:    longDesc,
		Example: example,
		Run: func(c *cobra.Command, args []string) {
			argsLenAtDash := c.ArgsLenAtDash()
			cmdutil.CheckErr(opts.Complete(c, args, argsLenAtDash))
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}
	cmd.Flags().BoolVar(&opts.DeploymentOnly, "deployment-only", false,
		"Only triage deployments in a given namespace")

	opts.Flags.AddFlags(cmd.Flags())

	return cmd
}

// Complete populate default values from KUBECONFIG file, sets up the clients
func (o *DoctorOptions) Complete(cmd *cobra.Command, args []string, argsLenAtDash int) error {

	o.Args = args
	if len(args) == 0 {
		log.Info("Going for a full scan as no flags are set!")
		o.FullScan = true
	}
	o.KubeCli = clientset

	var err error

	configLoader := o.Flags.ToRawKubeConfigLoader()

	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(o.Flags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	o.RESTClient, err = f.RESTClient()
	if err != nil {
		return err
	}
	log.Info("Retrieving necessary clientset for targeted k8s cluster.")
	o.CoreClient = clientset.CoreV1()

	fetchedNamespaces, _ := o.CoreClient.Namespaces().List(v1.ListOptions{})
	for _, i := range fetchedNamespaces.Items {
		o.FetchedNamespaces = append(o.FetchedNamespaces, i.GetName())
	}
	log.Info("")
	log.Info("Fetched namespaces: ", o.FetchedNamespaces)
	log.Info("")
	o.Config, err = configLoader.ClientConfig()
	if err != nil {
		return err
	}

	return nil
}

// Validate validate before the run that the namespace list cannot be empty(somehow?)
func (o *DoctorOptions) Validate() error {
	if len(o.FetchedNamespaces) == 0 {
		return errors.New("namespace must be specified properly!")
	}
	return nil
}

// Run doctor run
func (o *DoctorOptions) Run() error {

	// triage cluster crucial components starts
	log.Info("---")
	log.Info("Starting triage of cluster crucial component health checks.")
	componentsTriage, err := triage.TriageComponents(o.CoreClient)
	if err != nil {
		return err
	}
	if len(componentsTriage.Anomalies) == 0 {
		log.Info("Finished triage of cluster components, all clear!")
	} else {
		log.WithFields(log.Fields{"resource": componentsTriage.ResourceType, "Anomalies": componentsTriage.Anomalies}).Error(componentsTriage.AnomalyType)
	}
	// triage cluster crucial components ends

	// triage nodes stars
	log.Info("---")
	log.Info("Starting triage of nodes that form the cluster.")
	nodesTriage, err := triage.TriageNodes(o.CoreClient)
	if err != nil {
		return err
	}
	if len(nodesTriage.Anomalies) == 0 {
		log.Info("Nodes are looking good, no issues!")
	} else {
		log.WithFields(log.Fields{"resource": nodesTriage.ResourceType, "Anomalies": nodesTriage.Anomalies}).Warn(nodesTriage.AnomalyType)
	}
	// triage nodes ends

	// triage endpoints starts
	log.Info("---")
	log.Info("Starting triage of cluster-wide Endpoints resources.")

	endpointsTriage, err := triage.TriageEndpoints(o.CoreClient)
	if err != nil {
		return err
	}
	if len(endpointsTriage.Anomalies) == 0 {
		log.Info("Finished triage of Endpoints resources, all clear!")
	} else {
		log.WithFields(log.Fields{"resource": endpointsTriage.ResourceType, "Anomalies": endpointsTriage.Anomalies}).Warn(endpointsTriage.AnomalyType)
	}
	// triage endpoints ends

	// triage pvc starts
	log.Info("---")
	log.Info("Starting triage of cluster-wide pvc resources.")
	pvcTriage, err := triage.TriagePVC(o.CoreClient)
	if err != nil {
		return err
	}
	if len(pvcTriage.Anomalies) == 0 {
		log.Info("Finished triage of pvc resources, all clear!")
	} else {
		log.WithFields(log.Fields{"resource": pvcTriage.ResourceType, "Anomalies": pvcTriage.Anomalies}).Warn(pvcTriage.AnomalyType)
	}
	// triage pvc ends

	// triage pv starts
	log.Info("---")
	log.Info("Starting triage of cluster-wide pv resources.")
	pvTriage, err := triage.TriagePV(o.CoreClient)
	if err != nil {
		return err
	}
	if len(pvTriage.Anomalies) == 0 {
		log.Info("Finished triage of pv resources, all clear!")
	} else {
		log.WithFields(log.Fields{"resource": pvTriage.ResourceType, "Anomalies": pvTriage.Anomalies}).Warn(pvTriage.AnomalyType)
	}
	// triage pv ends

	// triage deployments starts
	log.Info("---")
	log.Info("Starting triage of deployment resources across cluster")
	for _, ns := range o.FetchedNamespaces {
		odeploymentTriage, err := triage.OrphanedDeployments(o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(odeploymentTriage.Anomalies) > 0 {
			log.WithFields(log.Fields{"resource": odeploymentTriage.ResourceType, "Anomalies": odeploymentTriage.Anomalies}).Warn(odeploymentTriage.AnomalyType)
		}

		ldeploymentTriage, err := triage.LeftOverDeployments(o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(ldeploymentTriage.Anomalies) > 0 {
			log.WithFields(log.Fields{"resource": ldeploymentTriage.ResourceType, "Anomalies": ldeploymentTriage.Anomalies}).Warn(ldeploymentTriage.AnomalyType)
		}

	}

	// triage deployments ends

	// triage replicasets starts
	log.Info("---")
	log.Info("Starting triage of replicasets resources across cluster")
	for _, ns := range o.FetchedNamespaces {
		orsTriage, err := triage.OrphanedReplicaSet(o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(orsTriage.Anomalies) > 0 {
			log.WithFields(log.Fields{"resource": orsTriage.ResourceType, "Anomalies": orsTriage.Anomalies}).Warn(orsTriage.AnomalyType)
		}
		lrsTriage, err := triage.LeftOverReplicaSet(o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(lrsTriage.Anomalies) > 0 {
			log.WithFields(log.Fields{"resource": lrsTriage.ResourceType, "Anomalies": lrsTriage.Anomalies}).Warn(lrsTriage.AnomalyType)
		}
	}

	// triage replicasets ends

	return nil

}
