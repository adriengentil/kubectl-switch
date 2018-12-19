package main

import (
	"errors"
	"fmt"
	"github.com/tjamet/kubectl-switch/kubectl"
	"github.com/tjamet/kubectl-switch/osext"
	"github.com/tjamet/kubectl-switch/server"
	"github.com/tjamet/kubectl-switch/update"
	"io"
	"os"

	"github.com/spf13/cobra"
	utilflag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var exit = os.Exit

func run(rg server.RestConfigGetter) {
	v := server.GetVersionFromConfig(rg)
	if !kubectl.Installed(v) {
		err := kubectl.Download(v)
		if err != nil {
			fmt.Printf("Failed to download kubectl version %s: %v\n", v, err.Error())
			exit(1)
		}
	}
	exit(kubectl.Exec(v, os.Args[1:]...))
}

type nopWriter struct{}

func (n nopWriter) Write(a []byte) (int, error) {
	return len(a), nil
}

type dummyPatcher struct{}

func (d dummyPatcher) Patch(old io.Reader, new io.Writer, patch io.Reader) error {
	return errors.New("dummy error")
}

func main() {
	if os.Getenv("KUBECTLSWITCH_SELF_UPDATE") == "true" {
		exe, err := osext.Executable()
		if err == nil {
			update.DefaultGitHub.ConfirmAndUpdate(exe)
		}
	}

	cmds := &cobra.Command{}

	flags := cmds.PersistentFlags()
	flags.SetNormalizeFunc(utilflag.WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)

	kubeConfigFlags := genericclioptions.NewConfigFlags()
	kubeConfigFlags.AddFlags(flags)
	cmds.Run = func(cmd *cobra.Command, args []string) {
		run(kubeConfigFlags)
	}
	cmds.SetUsageFunc(func(*cobra.Command) error {
		run(kubeConfigFlags)
		return nil
	})
	cmds.SetOutput(nopWriter{})
	cmds.Execute()
}
