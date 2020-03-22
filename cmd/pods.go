package cmd

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewPodsCommand(b Builder) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "pods [flags] [context]",
		Short: "Get pod names from Kube cluster specified by the context",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RunCommon(cmd); err != nil {
				return err
			}
			return RunPods(b, cmd, args)
		},
	}

	AddCommonFlags(cmd)
	cmd.Flags().StringP("namespace", "n", "", "Retrieve pods for a specific namespace (default is all)")

	return cmd
}

func RunPods(b Builder, cmd *cobra.Command, args []string) error {
	var context string

	kubeConfigFile, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		return err
	}
	kubeConfig, err := clientcmd.LoadFromFile(kubeConfigFile)
	if err != nil {
		return err
	}

	if len(args) > 1 {
		return errors.New("only 1 context can be specified")
	}
	if len(args) == 0 {
		// use the active context from kubeconfig
		context = kubeConfig.CurrentContext
	} else {
		// check that the specified context exists, if so use it
		if _, ok := kubeConfig.Clusters[args[0]]; ok {
			context = args[0]
		} else {
			return fmt.Errorf("unknown context: %s", args[0])
		}
	}

	log.Debugf("using context: %s", context)

	bind, err := GetBind(cmd)
	if err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	client, err := b.WatchClient(bind)
	if err != nil {
		return fmt.Errorf("could not create client to autocli: %s", err)
	}

	return nil
}
