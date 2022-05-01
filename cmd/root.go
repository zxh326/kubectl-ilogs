package main

import (
	"github.com/zxh326/kubectl-ilogs/pkg/cmd"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-ns", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewCmdIlogs(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
