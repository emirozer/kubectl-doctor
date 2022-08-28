package client

import (
	"flag"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	// utilities for kubernetes integration
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func init() {

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	// Only log the info severity or above.
	log.SetLevel(log.InfoLevel)
}

// InitClient - Kubernetes Client
func InitClient() *kubernetes.Clientset {
	// determine which kubeconfig to use
	var kubeconfig *string
	var kubeconfigbase string
	var kubecontext *string

	kubeconfigFromEnv, err := tryGetKubeConfigFromEnvVar()
	if err != nil {
		log.Warn(err.Error())
	} else {
		return kubeconfigFromEnv
	}

	// creating a client from env didn't work try auto discover
	if home := homeDir(); home != "" {
		kubeconfigbase = filepath.Join(home, ".kube", "config")
	}

	kubeconfig = flag.String(
		"kubeconfig",
		kubeconfigbase,
		"(optional) absolute path to the kubeconfig file",
	)
	kubecontext = flag.String(
		"context",
		"",
		"(optional) name of kube context",
	)

	flag.Parse()

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: *kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: *kubecontext,
		}).ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	csBackup, err := getClientSetFromConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return csBackup
}

func tryGetKubeConfigFromEnvVar() (*kubernetes.Clientset, error) {
	env_config := os.Getenv("KUBECONFIG")
	var delimeter string

	if env_config != "" {
		if runtime.GOOS == "windows" {
			delimeter = ";"
		} else {
			delimeter = ":"
		}
		if strings.Contains(env_config, delimeter) {
			// kubeconfig env var is a list, handle that
			log.WithFields(log.Fields{
				"kubeconfiglist": clientcmd.NewDefaultClientConfigLoadingRules().Precedence,
			}).Warn("discovered a list of kubeconfigs & will respect current-context!")
			for _, i := range clientcmd.NewDefaultClientConfigLoadingRules().Precedence {
				// if a problem occurs here it generally means that we are trying to build a client
				// that does not respect the current-context so if that happens just pass that kubeconfig file
				config, err := clientcmd.BuildConfigFromFlags("", i)
				if err != nil {
					continue
				}
				cs, err := getClientSetFromConfig(config)
				if err != nil {
					continue
				}
				return cs, nil
			}

		}

		config, err := clientcmd.BuildConfigFromFlags("", env_config)
		cs, err := getClientSetFromConfig(config)
		if err != nil {
			return nil, err
		}
		return cs, nil

	} else {
		return nil, errors.New("KUBECONFIG env var not found falling back to auto discovery!")
	}
}

func getClientSetFromConfig(config *restclient.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	// windows case
	return os.Getenv("USERPROFILE")
}
