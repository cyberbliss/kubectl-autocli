package cmd

import (
	"autocli/model"
	"autocli/service"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	strUtil "github.com/agrison/go-commons-lang/stringUtils"
)

type Builder interface {
	StdOut() io.Writer
	PodCompleter(in prompt.Document) []prompt.Suggest
	PopulateSuggestions(resources *[]model.KubeResource)
	PopulateContextSuggestions(source map[string][][]string)
	KubeClient(clients map[string]kubernetes.Interface) service.KubeClient
	WatchCache() *WatchCache
	WatchClient(address string) (WatchClient, error)
	Serve(l net.Listener, c *WatchCache) error
	SetCmdOptions(cmdoptions cmdOptions)
}

type DefaultBuilder struct {
	Streams     genericclioptions.IOStreams
	suggestions []prompt.Suggest
	contextSuggestions map[string][]prompt.Suggest
	cmdOptions cmdOptions
}

type cmdOptions func() []prompt.Suggest

func NewBuilder() Builder {
	return &DefaultBuilder{
		Streams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		cmdOptions: getDefaultOptions,
	}
}

func (b *DefaultBuilder) StdOut() io.Writer {
	return b.Streams.Out
}

func (b *DefaultBuilder) SetCmdOptions(options cmdOptions) {
	b.cmdOptions = options
}

func (b *DefaultBuilder) PopulateSuggestions(resources *[]model.KubeResource) {
	sort.Sort(model.ByKindNSName(*resources))
	s := make([]prompt.Suggest, 0)
	for _, res := range *resources {
		var text string
		if strUtil.IsBlank(res.Namespace) {
			text = res.Name
		} else {
			text = fmt.Sprintf("%s [%s]", res.Name, res.Namespace)
		}
		s = append(s, prompt.Suggest{
			Text:        text,
			Description: res.Status,
		})
	}

	b.suggestions = s
}

func (b *DefaultBuilder) PopulateContextSuggestions(source map[string][][]string) {
	target := make(map[string][]prompt.Suggest)
	for k,v := range source {
		s := make([]prompt.Suggest, 0)
		for _, val := range v {
			s = append(s, prompt.Suggest{
				Text: val[0],
				Description: val[1],
			})
		}
		target[k] = s
	}

	b.contextSuggestions = target
}

func (b *DefaultBuilder) PodCompleter(in prompt.Document) []prompt.Suggest {
	currText := in.CurrentLineBeforeCursor()
	//log.Debugf("pc: %s",currText)
	podChosen := b.isNameSelected(currText)

	//if a Pod name has already been selected or text then a space then don't display them again
	//and determine what other options to display: extra flags for example
	if podChosen || isAlreadyText(currText) {
		//wbc := in.GetWordBeforeCursor()
		//log.Debugf("ct: %s", currText)
		if strings.Contains(currText, "--container") {
			//`log.Debug("wbc true")
			pn := StringBefore(currText, " [")
			ns := StringBetween(currText, "[", "]")
			key := fmt.Sprintf("%s-%s", pn, ns)
			return b.contextSuggestions[key]
		}
		return b.cmdOptions()
	}

	return prompt.FilterContains(b.suggestions, in.GetWordBeforeCursor(), true)
}

func (b *DefaultBuilder) KubeClient(clients map[string]kubernetes.Interface) service.KubeClient {
	return service.NewKubeClient(clients)
}

func (b *DefaultBuilder) WatchCache() *WatchCache {
	return NewWatchCache()
}

func (b *DefaultBuilder) WatchClient(address string) (WatchClient, error) {
	return NewWatchClient(address, reflect.TypeOf(b).String(), "")
}

func (b *DefaultBuilder) Serve(l net.Listener, cache *WatchCache) error {
	rpc.RegisterName(reflect.TypeOf(b).String(),cache)
	rpc.HandleHTTP()
	return http.Serve(l, nil)
}

func (b *DefaultBuilder) isNameSelected(text string) bool {
	res := false
	for _, s := range b.suggestions {
		if strings.Contains(text, s.Text) {
			res = true
			break
		}
	}
	return res
}

func isAlreadyText(text string) bool {
	if strUtil.IsBlank(text) {
		return false
	}
	if strUtil.EndsWith(text, " ") {
		return true
	}

	return false
}

func getGetOptions() []prompt.Suggest {
	options := []prompt.Suggest {
		{Text: "--output json", Description: "Output manifest in json format"},
		{Text: "--output yaml", Description: "Output manifest in yaml format"},
		{Text: "--output wide", Description: "Output more details"},
		{Text: "--watch", Description: "After listing/getting the requested object, watch for changes"},
	}

	return options
}

func getLogOptions() []prompt.Suggest {
	options := []prompt.Suggest{
		{Text: "--all-containers", Description: "Get all containers' logs in the pod"},
		{Text: "--container", Description: "Get logs for specific container"},
		{Text: "--follow", Description: "Specify if the logs should be streamed"},
		{Text: "--prefix",Description: "Prefix each log line with the log source (pod name and container name)"},
		{Text: "--previous", Description: "Print the logs for the previous instance of the container in a pod if it exists"},
		{Text: "--timestamps", Description: "Include timestamps on each line in the log output"},
	}

	return options
}

func getDefaultOptions() []prompt.Suggest {
	return []prompt.Suggest{}
}