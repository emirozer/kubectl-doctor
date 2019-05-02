package plugin

import (
	"fmt"
	"github.com/emirozer/kubectl-doctor/pkg/client"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
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
	clientset = client.InitClient()
}

// DoctorOptions specify what the doctor is going to do
type DoctorOptions struct {
	// target namespace to scan
	Namespace string

	// Doctor options
	DeploymentOnly bool
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
		Use:                   "doctor [-n NAMESPACE] -- COMMAND [args...]",
		DisableFlagsInUseLine: true,
		Short:                 "start triage for current targeted kubernetes cluster",
		Long:                  longDesc,
		Example:               example,
		Run: func(c *cobra.Command, args []string) {
			argsLenAtDash := c.ArgsLenAtDash()
			cmdutil.CheckErr(opts.Complete(c, args, argsLenAtDash))
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}
	cmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "",
		"Target namespace to triage, defaults to the first namespace in cluster")
	cmd.Flags().BoolVar(&opts.DeploymentOnly, "deployment-only", false,
		"Only triage deployments in a given namespace")
	opts.Flags.AddFlags(cmd.Flags())

	return cmd
}

// Complete populate default values from KUBECONFIG file
func (o *DoctorOptions) Complete(cmd *cobra.Command, args []string, argsLenAtDash int) error {
	o.Args = args
	if len(args) == 0 {
		fmt.Println("Going for a full scan as no flags are set")
	}
	o.KubeCli = clientset

	var err error
	configLoader := o.Flags.ToRawKubeConfigLoader()
	o.Namespace, _, err = configLoader.Namespace()
	if err != nil {
		return err
	}

	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(o.Flags.ToRESTConfig())
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	o.RESTClient, err = f.RESTClient()
	if err != nil {
		return err
	}

	o.Namespace = args[0]

	o.Config, err = configLoader.ClientConfig()
	if err != nil {
		return err
	}

	o.CoreClient = clientset.CoreV1()
	return nil
}

// Validate validate
func (o *DoctorOptions) Validate() error {
	if len(o.Namespace) == 0 {
		return fmt.Errorf("namespace must be specified properly!")
	}
	return nil
}

// Run run
func (o *DoctorOptions) Run() error {
	nodes, err := o.CoreClient.Nodes().Get(o.Namespace, v1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}
