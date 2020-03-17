package cmd

import (
	"autocli/service"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
	"strings"
)

func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "127.0.0.1", "The IP address where mirror is accessible")
	cmd.Flags().IntP("port", "p", 33033, "The port on which mirror is accessible")
	cmd.Flags().String("kubeconfig", "~/.kube/config", "Path to the kubeconfig file")
	cmd.Flags().BoolP("info", "i", false, "Enables verbose output")
	cmd.Flags().BoolP("verbose", "v", false, "Enables very verbose output")
}

func RunCommon(cmd *cobra.Command) error {
	// sort out the logging level
	isVerbose, err := cmd.Flags().GetBool("info")
	if err != nil {
		return err
	} else if isVerbose {
		service.EnableVerbose()
	}
	isVeryVerbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	} else if isVeryVerbose {
		service.EnableVeryVerbose()
	}

	kubeConfigFile, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		return err
	}
	// if the path to the user's kubeconfig file starts with a ~ then convert this to an absolute path
	if strings.HasPrefix(kubeConfigFile, "~/") {
		usr, _ := user.Current()
		homeDir := usr.HomeDir
		kubeConfigFile = filepath.Join(homeDir, kubeConfigFile[2:])
	}
	cmd.Flags().Set("kubeconfig", kubeConfigFile)

	return nil
}

func BuildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	log.Infof("context: %s, path: %s", context, kubeconfigPath)
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func GetBind(cmd *cobra.Command) (string, error) {
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		return "", err
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", address, port), nil
}
