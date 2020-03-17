package cmd

import (
	"autocli/service"
	"github.com/c-bata/go-prompt"
	"io"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"net"
	"net/http"
	"net/rpc"
	"os"
)
type Builder interface {
	StdOut() io.Writer
	PodCompleter(in prompt.Document) []prompt.Suggest
	KubeClient(clients map[string]kubernetes.Interface) service.KubeClient
	WatchCache() *service.WatchCache
	Serve(l net.Listener, c *service.WatchCache) error
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

func (b *DefaultBuilder) WatchCache() *service.WatchCache {
	return service.NewWatchCache()
}

func (b *DefaultBuilder) Serve(l net.Listener, cache *service.WatchCache) error {
	rpc.Register(cache)
	rpc.HandleHTTP()
	return http.Serve(l, nil)
}
