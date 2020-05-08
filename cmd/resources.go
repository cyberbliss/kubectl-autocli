package cmd

import (
	"autocli/model"
	"errors"
	"fmt"
	strUtil "github.com/agrison/go-commons-lang/stringUtils"
	"github.com/c-bata/go-prompt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/exec"
	"strings"
)

func NewResourcesCommand(b Builder) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "resources [flags] [context]",
		Short: "Generate & execute kubectl get/describe command for a specified resource type and name",
		Long: `
DESCRIPTION
	For the specified resource kind and Kube context use the TAB key to display
	the list of names and TAB or UP/DOWN to select one. Once you have selected
	a name either hit ENTER to execute 'kubectl get' for the selected resource
	or SPACE to select a further context sensitive argument from a list.

	Supported resources are:
		- pod, po, p (the default)
		- log, lo, l
		- node, no, n

	Example:
		'kubectl ag log' will display a prompt so you can select from a list of Pod names the logs you want to show 
`,
		Aliases: getResourcesAliases(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RunCommon(cmd); err != nil {
				return err
			}
			return RunResources(b, cmd, args)
		},
	}

	AddCommonFlags(cmd)
	cmd.Flags().StringP("namespace", "n", "", "Retrieve resources for a specific namespace (default is all)")

	return cmd
}

func RunResources(b Builder, cmd *cobra.Command, args []string) error {
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

	isVerbose := false
	isVeryVerbose := false
	logLvlArg := ""
	isVerbose, _ = cmd.Flags().GetBool("info")
	isVeryVerbose, _ = cmd.Flags().GetBool("verbose")
	if isVerbose || isVeryVerbose {
		logLvlArg = "--info"
	}

	client, err := b.WatchClient(bind, logLvlArg, kubeConfigFile, context)
	if err != nil {
		return err
	}

	wf := makeFilter(context, ns, realKind(cmd.CalledAs()))
	kr, err := client.Resources(wf)
	if err != nil {
		return err
	}
	b.PopulateSuggestions(&kr)

	kreq := deriveKindRequired(cmd.CalledAs())
	if kreq == "log" {
		populateContainerSuggestions(b, cmd.CalledAs(), &kr)
	}

	prefix := fmt.Sprintf("[%s] >> ", kreq)

	in := prompt.Input(prefix, b.PodCompleter,
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
		//executor(StringBefore(in, " ["), StringBetween(in, "[", "]"), StringAfter(in, "]"), cmd.CalledAs())
		executor(kreq, in)
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
func executor(kind, in string) {
	var cmdArgs []string
	input := strings.Split(in, " ")
	switch kind {
	case "pod","log":
		ns := StringBetween(input[1], "[", "]")
		if kind == "log" {
			cmdArgs = []string{"logs"}
		} else {
			cmdArgs = []string{"get",kind}
		}
		cmdArgs = append(cmdArgs, input[0],"--namespace",ns)
		if len(input) > 2 {
			cmdArgs = append(cmdArgs, input[2:]...)
		}
	case "node":
		cmdArgs = []string{"get", kind}
		cmdArgs = append(cmdArgs, input...)
	}
	//log.Debug(cmdArgs)
	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed with %s\n", err)
	}
}
//func executor(rn, ns, args, c string) {
//	//cmdArgs := []string{"get", "pod", rn, "-n", ns}
//	var cmdArgs []string
//	if strUtil.ContainsAny(c, getLogAliases()...) {
//		cmdArgs = append(cmdArgs, "logs")
//	} else {
//		cmdArgs = append(cmdArgs, "get", "pod")
//	}
//	cmdArgs = append(cmdArgs, rn, "-n", ns)
//
//	if strUtil.IsNotBlank(args) {
//		cmdArgs = append(cmdArgs, strings.Split(strUtil.Trim(args), " ")...)
//	}
//	cmd := exec.Command("kubectl", cmdArgs...)
//	cmd.Stderr = os.Stderr
//	cmd.Stdin = os.Stdin
//	cmd.Stdout = os.Stdout
//	err := cmd.Run()
//	if err != nil {
//		log.Fatalf("failed with %s\n", err)
//	}
//}

// once a pod name has been selected we want to provide context appropriate options
// dependent on whether this cmd was called with 'pod' (to get details on the pod) or 'log'
// (to get pod logs)
func deriveCmdOptions(cmd string) cmdOptions {
	kind := deriveKindRequired(cmd)
	switch kind {
	case "pod", "node":
		return getGetOptions
	case "log":
		return getLogOptions
	default:
		return getDefaultOptions
	}
}

func realKind(cmd string) string {
	srcKind := deriveKindRequired(cmd)
	if srcKind == "log" {
		return "pod"
	}

	return srcKind
}

func deriveKindRequired(cmd string) string {
	switch {
	case strUtil.ContainsAny(cmd, getPodAliases()...):
		return "pod"
	case strUtil.ContainsAny(cmd, getNodeAliases()...):
		return "node"
	case strUtil.ContainsAny(cmd, getLogAliases()...):
		return "log"
	default:
		return ""
	}
}

func getPodAliases() []string {
	return []string{"pod","p", "po"}
}

func getLogAliases() []string {
	return []string{"logs", "log", "lo", "l"}
}

func getNodeAliases() []string {
	return []string{"node", "no", "n"}
}

func populateContainerSuggestions(b Builder, cmd string, resources *[]model.KubeResource) {
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

func getResourcesAliases() []string {
	var aliases []string
	aliases = append(aliases, getPodAliases()...)
	aliases = append(aliases, getLogAliases()...)
	aliases = append(aliases, getNodeAliases()...)
	return aliases
}

