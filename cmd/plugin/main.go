package main

import (
	"os"

	"github.com/emirozer/kubectl-doctor/pkg/plugin"
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-doctor", pflag.ExitOnError)
	pflag.CommandLine = flags

	// bypass to DoctorCmd
	cmd := plugin.NewDoctorCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
