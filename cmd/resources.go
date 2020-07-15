package cmd

import (
	"autocli/model"
	"autocli/service"
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
	If a context is not specified then the active context from kubeconfig will be used.

	Supported resources are:
		- pod, po, p (the default)
		- log, lo, l
		- node, no, n
		- ssh (to exec into the selected Pod)

	Example:
		'kubectl ac log' will display a prompt so you can select from a list of Pod names the logs you want to show 
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
	cmd.Flags().Bool("setproxy", true, "If true then set the HTTPS_PROXY env var to the kube context's proxy-url value (if available) before executing kubectl. This is only relevant if a proxy is required to access the Kube Master AND kubectl version is < v1.19")

	return cmd
}

func RunResources(b Builder, cmd *cobra.Command, args []string) error {
	var context, proxyURL string

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
		if strUtil.IsBlank(context) {
			return fmt.Errorf("couldn't determine active context; please specify one")
		}
		proxyURL = kubeConfig.Clusters[context].ProxyURL
	} else {
		// check that the specified context exists, if so use it
		if cluster, ok := kubeConfig.Clusters[args[0]]; ok {
			context = args[0]
			proxyURL = cluster.ProxyURL
		} else {
			return fmt.Errorf("unknown context: %s", args[0])
		}
	}

	if val, _ := cmd.Flags().GetBool("setproxy"); !val {
		proxyURL = ""
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
	if kreq == "log" || kreq == "ssh" {
		populateContainerSuggestions(b, cmd.CalledAs(), &kr)
	}

	prefix := fmt.Sprintf("[%s] >> ", kreq)
	writer := service.NewStdoutWriter()
	in := prompt.Input(prefix, b.PodCompleter,
		prompt.OptionWriter(writer),
		prompt.OptionShowCompletionAtStart(),
		// Set the colours for the prompt and suggestions
		prompt.OptionPrefixTextColor(service.Themes["light"].OptionPrefixTextColor),
		prompt.OptionPrefixBackgroundColor(service.Themes["light"].OptionPrefixBackgroundColor),
		prompt.OptionPreviewSuggestionBGColor(service.Themes["light"].OptionPreviewSuggestionBGColor),
		prompt.OptionPreviewSuggestionTextColor(service.Themes["light"].OptionPreviewSuggestionTextColor),
		prompt.OptionInputBGColor(service.Themes["light"].OptionInputBGColor),
		prompt.OptionInputTextColor(service.Themes["light"].OptionInputTextColor),
		prompt.OptionScrollbarBGColor(service.Themes["light"].OptionScrollbarBGColor),
		prompt.OptionScrollbarThumbColor(service.Themes["light"].OptionScrollbarThumbColor),
		prompt.OptionSelectedSuggestionBGColor(service.Themes["light"].OptionSelectedSuggestionBGColor),
		prompt.OptionSelectedSuggestionTextColor(service.Themes["light"].OptionSelectedSuggestionTextColor),

		prompt.OptionDescriptionBGColor(service.Themes["light"].OptionDescriptionBGColor),
		prompt.OptionDescriptionTextColor(service.Themes["light"].OptionDescriptionTextColor),

		prompt.OptionSelectedDescriptionBGColor(service.Themes["light"].OptionSelectedDescriptionBGColor),
		prompt.OptionSelectedDescriptionTextColor(service.Themes["light"].OptionSelectedDescriptionTextColor),

		prompt.OptionSuggestionTextColor(service.Themes["light"].OptionSuggestionTextColor),
		prompt.OptionSuggestionBGColor(service.Themes["light"].OptionSuggestionBGColor))

	in = strings.TrimSpace(in)
	log.Debugf("Your input: %s", in)
	if strUtil.IsNotBlank(in) {
		executor(context, kreq, in, proxyURL)
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
func executor(ctx, kind, in, proxyURL string) {

	var (
		cmdArgs        []string
		getOrDescribe  string
		sanitisedInput []string
	)

	input := strings.Split(in, " ")

	switch {
	case Contains(input, "@describe"):
		getOrDescribe = "describe"
		//need to remove all elements in the input slice to prevent kubectl errors (apart from 1st two: pod name & namespace)
		sanitisedInput = input[:2]
	default:
		getOrDescribe = "get"
		sanitisedInput = input
	}

	switch kind {
	case "pod", "log":
		ns := StringBetween(sanitisedInput[1], "[", "]")
		if kind == "log" {
			cmdArgs = []string{"logs"}
		} else {
			cmdArgs = []string{getOrDescribe, kind}
		}
		cmdArgs = append(cmdArgs, sanitisedInput[0], "--namespace", ns)
		if len(sanitisedInput) > 2 {
			cmdArgs = append(cmdArgs, sanitisedInput[2:]...)
		}
		cmdArgs = append(cmdArgs, "--context", fmt.Sprintf("%s", ctx))
	case "ssh":
		ns := StringBetween(sanitisedInput[1], "[", "]")
		cmdArgs = []string{"exec", "-ti", sanitisedInput[0], "--namespace", ns, "--context", fmt.Sprintf("%s", ctx)}

		// check if a container has been specified, if so add that to the exec command
		if len(sanitisedInput) == 4 {
			cmdArgs = append(cmdArgs, sanitisedInput[2:]...)
		}

		// the shell command always needs to go last
		cmdArgs = append(cmdArgs, "--", "sh")
	case "node":
		cmdArgs = []string{getOrDescribe, kind}
		cmdArgs = append(cmdArgs, sanitisedInput...)
		cmdArgs = append(cmdArgs, "--context", fmt.Sprintf("%s", ctx))
	}

	log.Debug(cmdArgs)
	cmd := exec.Command("kubectl", cmdArgs...)
	if proxyURL != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "HTTPS_PROXY="+proxyURL)
	}
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
	kind := deriveKindRequired(cmd)
	switch kind {
	case "pod":
		return getPodGetOptions
	case "node":
		return getNodeGetOptions
	case "log":
		return getLogOptions
	case "ssh":
		return getSSHOptions
	default:
		return getDefaultOptions
	}
}

func realKind(cmd string) string {
	srcKind := deriveKindRequired(cmd)
	if srcKind == "log" || srcKind == "ssh" {
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
	case strUtil.ContainsAny(cmd, getSSHAliases()...):
		return "ssh"
	case cmd == "resources":
		return "pod"
	default:
		return ""
	}
}

func getPodAliases() []string {
	return []string{"pod", "p", "po"}
}

func getLogAliases() []string {
	return []string{"logs", "log", "lo", "l"}
}

func getNodeAliases() []string {
	return []string{"node", "no", "n"}
}

func getSSHAliases() []string {
	return []string{"ssh"}
}

func populateContainerSuggestions(b Builder, cmd string, resources *[]model.KubeResource) {
	cMap := make(map[string][][]string)
	for _, res := range *resources {
		key := fmt.Sprintf("%s-%s", res.Name, res.Namespace)
		value := make([][]string, 0)
		for _, v := range res.ContainerNames {
			value = append(value, []string{v.Name, v.Type})
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
	aliases = append(aliases, getSSHAliases()...)
	return aliases
}
