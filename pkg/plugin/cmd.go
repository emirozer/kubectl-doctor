package plugin

import (
	"github.com/cheggaaa/pb/v3"
	"github.com/emirozer/kubectl-doctor/pkg/client"
	"github.com/emirozer/kubectl-doctor/pkg/triage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
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

// Complete populate default values from KUBECONFIG file
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
	log.Info("Retrieving CoreV1 Client for targeted cluster.")
	o.CoreClient = clientset.CoreV1()

	fetchedNamespaces, _ := o.CoreClient.Namespaces().List(v1.ListOptions{})
	for _, i := range fetchedNamespaces.Items {
		o.FetchedNamespaces = append(o.FetchedNamespaces, i.GetName())
	}

	log.Info("Fetched namespaces: {}", o.FetchedNamespaces)

	o.Config, err = configLoader.ClientConfig()
	if err != nil {
		return err
	}

	return nil
}

// Validate validate
func (o *DoctorOptions) Validate() error {
	if len(o.FetchedNamespaces) == 0 {
		return errors.New("namespace must be specified properly!")
	}
	return nil
}

// Run run
func (o *DoctorOptions) Run() error {

	// triage endpoints starts
	log.Info("Starting triage of cluster-wide Endpoints resources.")
	endpoints, err := o.CoreClient.Endpoints("").List(v1.ListOptions{})
	if err != nil {
		return err
	}
	bar := pb.StartNew(len(endpoints.Items))
	listOfTriages := make([]string, 0)
	for _, i := range endpoints.Items {
		bar.Increment()
		time.Sleep(time.Millisecond)
		if len(i.Subsets) == 0 {
			listOfTriages = append(listOfTriages, i.GetName())
		}
	}
	endpointsTriage := triage.NewTriage("Endpoints", "Found orphaned endpoints!", listOfTriages)
	bar.Finish()
	if len(listOfTriages) == 0 {
		log.Info("Finished triage of Endpoints resources, all clear!")
	} else {
		log.WithFields(log.Fields{"resource": endpointsTriage.ResourceType, "Anomalies": endpointsTriage.Anomalies}).Warn(endpointsTriage.AnomalyType)
	}

	// triage endpoints ends

	return nil
}
