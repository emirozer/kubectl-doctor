package client

import (
	"flag"
	"github.com/pkg/errors"
	"os"
	"path/filepath"

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
	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}


// InitClient - Kubernetes Client
func InitClient() *kubernetes.Clientset {
	// determine which kubeconfig to use
	var kubeconfig *string
	var kubeconfigbase string

	kubeconfigFromEnv, err := tryGetKubeConfigFromEnvVar()
	if err!= nil {
		log.Warn(err.Error())
	} else {
		return getClientSetFromConfig(kubeconfigFromEnv)
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
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return getClientSetFromConfig(config)
}

func tryGetKubeConfigFromEnvVar() (*restclient.Config, error) {
	env_config := os.Getenv("KUBECONFIG")

	if env_config != "" {
		config, err := clientcmd.BuildConfigFromFlags("", env_config)
		if err != nil {
			panic(err.Error())
		}
		return config, nil
	} else {
		return  nil, errors.New("KUBECONFIG env var not found falling back to auto discovery!")
	}
}

func getClientSetFromConfig(config *restclient.Config) (*kubernetes.Clientset){
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	// windows case
	return os.Getenv("USERPROFILE")
}