package cmd

import (
	"autocli/model"
	"autocli/service"
	"fmt"
	"github.com/c-bata/go-prompt"
	log "github.com/sirupsen/logrus"
	"io"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"reflect"
	"sync"
)

type TestBuilder struct {
	Streams     genericclioptions.IOStreams
	suggestions []prompt.Suggest
	cmdOptions cmdOptions
	testKubeClient *TestKubeClient
}

func NewTestBuilder() Builder {
	return &TestBuilder{
		Streams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		cmdOptions: getDefaultOptions,
	}
}

func (t *TestBuilder) StdOut() io.Writer {
	return t.Streams.Out
}

func (t *TestBuilder) PodCompleter(in prompt.Document) []prompt.Suggest {
	panic("implement me")
}

func (t *TestBuilder) PopulateSuggestions(resources *[]model.KubeResource) {
	panic("implement me")
}

func (t *TestBuilder) KubeClient(clients map[string]kubernetes.Interface) service.KubeClient {
	testClients := map[string]kubernetes.Interface{}
	for key := range clients {
		testClients[key] = testclient.NewSimpleClientset()
	}
	t.testKubeClient = &TestKubeClient{
		clients: testClients,
		watchObjectLock: &sync.RWMutex{},
		watchObjectHits: map[string]int{},
	}

	return t.testKubeClient
}

func (t *TestBuilder) WatchCache() *WatchCache {
	return NewWatchCache()
}

func (t *TestBuilder) WatchClient(address string) (WatchClient, error) {
	return NewWatchClient(address, reflect.TypeOf(t).String(), "")
}

func (t *TestBuilder) Serve(l net.Listener, c *WatchCache) error {

	fmt.Println("in TestBuilder.Serve" + reflect.TypeOf(t).String())
	rpc.RegisterName(reflect.TypeOf(t).String(),c)
	rpc.HandleHTTP()


	return http.Serve(l, nil)

}

func (t *TestBuilder) SetCmdOptions(cmdoptions cmdOptions) {
	panic("implement me")
}

type TestKubeClient struct {
	clients map[string]kubernetes.Interface
	watchObjectHits  map[string]int
	watchObjectLock  *sync.RWMutex
}

func (t TestKubeClient) Ping(context string) error {
	return nil
}

func (t TestKubeClient) WatchResources(context, kind string, out chan *model.ResourceEvent) error {
	log.Debug("in WatchResources")
	t.watchObjectLock.Lock()
	t.watchObjectHits[kind] += 1
	t.watchObjectLock.Unlock()

	var evt model.ResourceEvent

	var podname, nsname string
	if context == "prod" {
		podname = "ns1-pod"
		nsname = "ns1"
	} else {
		podname = "ns2-pod"
		nsname = "ns2"
	}
	evt.Type = model.Added
	evt.Resource = &model.KubeResource{
		TypeMeta:   model.TypeMeta{Kind: "pod"},
		ResourceMeta: model.ResourceMeta{
			Name: podname,
			Namespace: nsname,
			Status: "Running",
		},
	}
	out <- &evt

	return nil
}

