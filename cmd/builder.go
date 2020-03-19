package cmd

import (
	"autocli/service"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/c-bata/go-prompt"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type Builder interface {
	StdOut() io.Writer
	PodCompleter(in prompt.Document) []prompt.Suggest
	KubeClient(clients map[string]kubernetes.Interface) service.KubeClient
	WatchCache() *WatchCache
	WatchClient(address string) (WatchClient, error)
	Serve(l net.Listener, c *WatchCache) error
}

type DefaultBuilder struct {
	Streams genericclioptions.IOStreams
	//suggestions []prompt.Suggest
	//PodCompleter prompt.Completer
}

func NewBuilder() Builder {
	return &DefaultBuilder{
		Streams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

func (b *DefaultBuilder) StdOut() io.Writer {
	return b.Streams.Out
}

func (b *DefaultBuilder) PodCompleter(in prompt.Document) []prompt.Suggest {
	s := service.GetPods()
	return prompt.FilterContains(s, in.GetWordBeforeCursor(), true)
}

func (b *DefaultBuilder) KubeClient(clients map[string]kubernetes.Interface) service.KubeClient {
	return service.NewKubeClient(clients)
}

func (b *DefaultBuilder) WatchCache() *WatchCache {
	return NewWatchCache()
}

func (b *DefaultBuilder) WatchClient(address string) (WatchClient, error) {
	return NewWatchClient(address)
}

func (b *DefaultBuilder) Serve(l net.Listener, cache *WatchCache) error {
	rpc.Register(cache)
	rpc.HandleHTTP()
	return http.Serve(l, nil)
}
