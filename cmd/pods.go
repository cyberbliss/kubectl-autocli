package cmd

import (
	"autocli/model"
	"errors"
	"fmt"
	"github.com/c-bata/go-prompt"
	"os"
	"os/exec"
	"strings"

	strUtil "github.com/agrison/go-commons-lang/stringUtils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewPodsCommand(b Builder) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "pods [flags] [context]",
		Short: "Get pod names from Kube cluster specified by the context (default is your active context)",
		Aliases: append(getPodAliases(), getLogAliases()...),
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

	b.SetCmdOptions(deriveCmdOptions(cmd.CalledAs()))

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

	ns, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	log.Debugf("using context: %s and namespace %s", context, ns)

	bind, err := GetBind(cmd)
	if err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	client, err := b.WatchClient(bind)
	if err != nil {
		return fmt.Errorf("could not create client to autocli: %s", err)
	}

	wf := makeFilter(context, ns,"pod")
	kr, err := client.Resources(wf)
	if err != nil {
		return err
	}
	b.PopulateSuggestions(&kr)
	populateContainerSuggestions(b, cmd.CalledAs(), &kr)

	in := prompt.Input(">>> ", b.PodCompleter,
		// Set the colours for the prompt and suggestions
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),

		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.DarkRed),

		prompt.OptionDescriptionBGColor(prompt.White),
		prompt.OptionDescriptionTextColor(prompt.DarkBlue),

		prompt.OptionSelectedDescriptionBGColor(prompt.LightGray),
		prompt.OptionSelectedDescriptionTextColor(prompt.Purple),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSuggestionBGColor(prompt.DarkBlue))

	log.Debugf("Your input: %s", in)
	if strUtil.IsNotBlank(in) {
		executor(StringBefore(in, " ["), StringBetween(in, "[", "]"), StringAfter(in, "]"), cmd.CalledAs())
	}
	return nil
}

func makeFilter(context, ns, kind string) WatchFilter {
	wf := WatchFilter{
		Context: context,
		Kind:    kind,
	}
	if kind == "node" {
		wf.Namespace = ""
	} else {
		wf.Namespace = ns
	}

	return wf
}

func executor(rn, ns, args, c string) {
	//cmdArgs := []string{"get", "pod", rn, "-n", ns}
	var cmdArgs []string
	if strUtil.ContainsAny(c, getLogAliases()...) {
		cmdArgs = append(cmdArgs, "logs")
	} else {
		cmdArgs = append(cmdArgs, "get", "pod")
	}
	cmdArgs = append(cmdArgs, rn, "-n", ns)

	if strUtil.IsNotBlank(args) {
		cmdArgs = append(cmdArgs, strings.Split(strUtil.Trim(args), " ")...)
	}
	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed with %s\n", err)
	}
}

// once a pod name has been selected we want to provide context appropriate options
// dependent on whether this cmd was called with 'pod' (to get details on the pod) or 'log'
// (to get pod logs)
func deriveCmdOptions(cmd string) cmdOptions {
	if strUtil.ContainsAny(cmd, getPodAliases()...) {
		return getPodOptions
	}
	if strUtil.ContainsAny(cmd, getLogAliases()...) {
		return getLogOptions
	}

	return getDefaultOptions
}

func getPodAliases() []string {
	return []string{"pods","pod","p", "po"}
}

func getLogAliases() []string {
	return []string{"logs", "log", "lo", "l"}
}

func populateContainerSuggestions(b Builder, cmd string, resources *[]model.KubeResource) {
	if strUtil.ContainsAny(cmd, getPodAliases()...) {
		return
	}
	cMap := make(map[string][][]string)
	for _, res := range *resources {
		key := fmt.Sprintf("%s-%s", res.Name, res.Namespace)
		value := make([][]string,0)
		for _, v := range res.ContainerNames {
			value = append(value,[]string{v.Name, v.Type})
		}
		//value := res.ContainerNames
		cMap[key] = value
	}

	b.PopulateContextSuggestions(cMap)
}
