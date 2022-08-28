package plugin

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-semver/semver"
	"github.com/emirozer/kubectl-doctor/pkg/client"
	"github.com/emirozer/kubectl-doctor/pkg/triage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const (
	example = `
	# triage everything in the cluster
	kubectl doctor
`
	longDesc = `
    kubectl-doctor plugin will scan the given k8s cluster for any kind of anomalies and reports back to its user.
    example anomalies:
        * deployments that are older than 30d with 0 available,
        * deployments that do not have minimum availability,
        * kubernetes nodes cpu usage or memory usage too high. or too low to report scaledown possiblity
`

	usageError = "expects no flags .. 'doctor' for doctor command"
)

const K8S_CLIENT_VERSION = "11.0.0"

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
	Context        context.Context
}

// NewDoctorOptions new doctor options initializer
func NewDoctorOptions() *DoctorOptions {
	return &DoctorOptions{
		Flags:   genericclioptions.NewConfigFlags(true),
		Context: context.Background(),
	}
}

// NewDoctorCmd returns a cobra command wrapping DoctorOptions
func NewDoctorCmd() *cobra.Command {

	opts := NewDoctorOptions()

	cmd := &cobra.Command{
		Use:     "doctor",
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

	fetchedNamespaces, _ := o.CoreClient.Namespaces().List(o.Context, v1.ListOptions{})
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
		return errors.New("namespace must be specified/retrieved properly!")
	}

	// compat check
	serverVersion, err := o.KubeCli.ServerVersion()
	if err != nil {
		return err
	}
	// vx.x.x to x.x.x
	serverVersionStr := serverVersion.String()[1:]

	serverSemVer := semver.New(serverVersionStr)
	if serverSemVer.Major != 1 || serverSemVer.Minor != 14 {
		log.Warn("doctor's client-go version: " + K8S_CLIENT_VERSION + " is not fully compatible with your k8s server version: " + serverVersionStr)
		log.Warn("https://github.com/kubernetes/client-go#compatibility-matrix")
		log.Warn("doctor run will be based on best-effort delivery")
	}
	return nil
}

// Run doctor run
func (o *DoctorOptions) Run() error {
	// report setup
	report := make(map[interface{}][]*triage.Triage)
	report["TriageReport"] = make([]*triage.Triage, 0)

	// triage cluster crucial components starts
	log.Info("Starting triage of cluster crucial component health checks.")
	componentsTriage, err := triage.TriageComponents(o.Context, o.CoreClient)
	if err != nil {
		return err
	}
	if len(componentsTriage.Anomalies) > 0 {
		report["TriageReport"] = append(report["TriageReport"], componentsTriage)
	}
	// triage cluster crucial components ends

	// triage nodes stars
	log.Info("Starting triage of nodes that form the cluster.")
	nodesTriage, err := triage.TriageNodes(o.Context, o.CoreClient)
	if err != nil {
		return err
	}
	if len(nodesTriage.Anomalies) > 0 {
		report["TriageReport"] = append(report["TriageReport"], nodesTriage)
	}
	// triage nodes ends

	// triage pvc starts
	log.Info("Starting triage of cluster-wide pvc resources.")
	pvcTriage, err := triage.TriagePVC(o.Context, o.CoreClient)
	if err != nil {
		return err
	}
	if len(pvcTriage.Anomalies) > 0 {
		report["TriageReport"] = append(report["TriageReport"], pvcTriage)
	}
	// triage pvc ends

	// triage pv starts
	log.Info("Starting triage of cluster-wide pv resources.")
	pvTriage, err := triage.TriagePV(o.Context, o.CoreClient)
	if err != nil {
		return err
	}
	if len(pvTriage.Anomalies) > 0 {
		report["TriageReport"] = append(report["TriageReport"], pvTriage)
	}
	// triage pv ends

	// triage ns starts
	log.Info("Starting triage of namespace resources.")
	nsTriage, err := triage.TerminatingNamespaces(o.Context, o.KubeCli)
	if err != nil {
		return err
	}
	if len(nsTriage.Anomalies) > 0 {
		report["TriageReport"] = append(report["TriageReport"], pvTriage)
	}
	// triage ns ends

	// triage deployments starts
	log.Info("Starting triage of deployment resources across cluster")
	for _, ns := range o.FetchedNamespaces {
		odeploymentTriage, err := triage.OrphanedDeployments(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(odeploymentTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], odeploymentTriage)
		}

		ldeploymentTriage, err := triage.LeftOverDeployments(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(ldeploymentTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], ldeploymentTriage)
		}

	}

	// triage deployments ends

	// triage replicasets starts
	log.Info("Starting triage of replicasets resources across cluster")
	for _, ns := range o.FetchedNamespaces {
		orsTriage, err := triage.OrphanedReplicaSet(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(orsTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], orsTriage)
		}
		lrsTriage, err := triage.LeftOverReplicaSet(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(lrsTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], lrsTriage)
		}
	}

	// triage replicasets ends

	// triage pods starts
	log.Info("Starting triage of pod resources across cluster")
	for _, ns := range o.FetchedNamespaces {
		standalonePodTriage, err := triage.StandalonePods(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(standalonePodTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], standalonePodTriage)
		}
		unreadyPodTriage, err := triage.UnreadyPods(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(unreadyPodTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], unreadyPodTriage)
		}
	}

	// triage pods ends

	// triage endpoints starts
	log.Info("Starting triage of endpoints resources across cluster.")
	for _, ns := range o.FetchedNamespaces {
		endpointsTriage, err := triage.TriageEndpoints(o.Context, o.CoreClient, ns)
		if err != nil {
			return err
		}
		if len(endpointsTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], endpointsTriage)
		}
	}

	// triage endpoints ends

	// triage jobs starts
	log.Info("Starting triage of cronjob resources across cluster")
	var jobsTriage *triage.Triage
	for _, ns := range o.FetchedNamespaces {
		jobsTriage, err = triage.LeftoverJobs(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(jobsTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], jobsTriage)
		}
	}
	// triage jobs end

	// triage ingresses starts
	log.Info("Starting triage of ingress resources across cluster")
	var ingressTriage *triage.Triage
	for _, ns := range o.FetchedNamespaces {
		ingressTriage, err = triage.LeftoverIngresses(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(ingressTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], ingressTriage)
		}
	}
	// triage ingresses ends

	// triage PodDisruptionBudgets starts
	log.Info("Starting triage of PodDisruptionBudget resources across cluster")
	var pdbTriage *triage.Triage
	for _, ns := range o.FetchedNamespaces {
		pdbTriage, err = triage.OrphanedDisruptionBudget(o.Context, o.KubeCli, ns)
		if err != nil {
			return err
		}
		if len(pdbTriage.Anomalies) > 0 {
			report["TriageReport"] = append(report["TriageReport"], pdbTriage)
		}
	}
	// triage PodDisruptionBudgets ends

	// yaml outputter
	if len(report["TriageReport"]) > 0 {
		log.Info("Triage report coming up in yaml format:")
		d, err := yaml.Marshal(&report)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Println("\n---\n", string(d))
	} else {
		log.Info("Triage finished, cluster all clear, no anomalies detected!")
	}

	return nil

}
