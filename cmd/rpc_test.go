package cmd

import (
	"autocli/model"
	v1 "k8s.io/api/core/v1"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	cache       *WatchCache
	watchClient WatchClient
	once        sync.Once
)

func setupRPC() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to bind: %v", err)
	}
	b := &DefaultBuilder{}
	cache := b.WatchCache()
	fillCache(cache)
	rpc.Register(cache)
	rpc.HandleHTTP()
	go http.Serve(l, nil)

	watchClient, err = b.WatchClient(l.Addr().String())
	if err != nil {
		log.Fatalf("Failed to create WatchClient: %v", err)
	}
}

func fillCache(c *WatchCache) {
	for _, ctx := range []string{"ctx1", "ctx2", "ctx3"} {
		for _, ns := range []string{"ns1", "ns2", "ns3"} {
			for _, kind := range []string{"pod", "service", "deployment"} {
				for _, name := range []string{"a", "b", "c"} {
					if c.resources[ctx] == nil {
						c.resources[ctx] = make([]model.KubeResource, 0)
					}

					r := model.KubeResource{
						model.TypeMeta{Kind: kind},
						model.ResourceMeta{Name: ctx + "-" + name, Namespace: ns, Status: string(v1.PodRunning)},
					}
					c.resources[ctx] = append(c.resources[ctx], r)
				}
			}
		}
	}

	for _, ctx := range []string{"ctx1", "ctx2"} {
		for _, ns := range []string{"ns1", "ns2"} {
			r := model.KubeResource{
				model.TypeMeta{"namespace"},
				model.ResourceMeta{Name: ctx + "-" + ns, Status: string(v1.PodRunning)},
			}
			c.resources[ctx] = append(c.resources[ctx], r)
		}
	}
}

func TestClientResources(t *testing.T) {
	once.Do(setupRPC)

	tests := []struct {
		filter   WatchFilter
		expected []model.KubeResource
		isError  bool
	}{
		{
			filter: WatchFilter{},
		},
		{
			filter: WatchFilter{
				Context:   "context_other",
				Namespace: "ns1",
				Kind:      "pod",
			},
			isError: true,
		},
		{
			filter: WatchFilter{
				Context:   "ctx1",
				Namespace: "ns_other",
				Kind:      "pod",
			},
		},
		{
			filter: WatchFilter{
				Context:   "ctx1",
				Namespace: "ns1",
				Kind:      "pod_other",
			},
		},
		{
			filter: WatchFilter{"CTX1", "ns1", "pod"},
			expected: []model.KubeResource{
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-c", "ns1", "", string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"ctx2", "NS1", "pod"},
			expected: []model.KubeResource{
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx2-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx2-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx2-c", "ns1", "", string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"ctx1", "ns2", "POD"},
			expected: []model.KubeResource{
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-a", "ns2", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-b", "ns2", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-c", "ns2", "", string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"ctx1", "ns1", "service"},
			expected: []model.KubeResource{
				{model.TypeMeta{"service"}, model.ResourceMeta{"ctx1-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"service"}, model.ResourceMeta{"ctx1-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"service"}, model.ResourceMeta{"ctx1-c", "ns1", "", string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"ctx1", "ns1", "deployment"},
			expected: []model.KubeResource{
				{model.TypeMeta{"deployment"}, model.ResourceMeta{"ctx1-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"deployment"}, model.ResourceMeta{"ctx1-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"deployment"}, model.ResourceMeta{"ctx1-c", "ns1", "",string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"", "ns1", "pod"},
			expected: []model.KubeResource{
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-c", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx2-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx2-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx2-c", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx3-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx3-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx3-c", "ns1", "", string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"ctx1", "", "pod"},
			expected: []model.KubeResource{
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-a", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-b", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-c", "ns1", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-a", "ns2", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-b", "ns2", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-c", "ns2", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-a", "ns3", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-b", "ns3", "", string(v1.PodRunning)}},
				{model.TypeMeta{"pod"}, model.ResourceMeta{"ctx1-c", "ns3", "", string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"ctx1", "should be ignored", "namespace"},
			expected: []model.KubeResource{
				{model.TypeMeta{"namespace"}, model.ResourceMeta{"ctx1-ns1", "", "", string(v1.PodRunning)}},
				{model.TypeMeta{"namespace"}, model.ResourceMeta{"ctx1-ns2", "", "", string(v1.PodRunning)}},
			},
		},
		{
			filter: WatchFilter{"", "should be ignored", "namespace"},
			expected: []model.KubeResource{
				{model.TypeMeta{"namespace"}, model.ResourceMeta{"ctx1-ns1", "", "", string(v1.PodRunning)}},
				{model.TypeMeta{"namespace"}, model.ResourceMeta{"ctx1-ns2", "", "", string(v1.PodRunning)}},
				{model.TypeMeta{"namespace"}, model.ResourceMeta{"ctx2-ns1", "", "", string(v1.PodRunning)}},
				{model.TypeMeta{"namespace"}, model.ResourceMeta{"ctx2-ns2", "", "", string(v1.PodRunning)}},
			},
		},
	}

	for _, test := range tests {
		actual, err := watchClient.Resources(test.filter)
		if !test.isError && err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		assert.Equal(t, test.expected, actual)
	}
}

func TestDeleteKubeObjects(t *testing.T) {
	c := NewWatchCache()
	s := "s"
	o1 := model.KubeResource{TypeMeta: model.TypeMeta{"x"}, ResourceMeta: model.ResourceMeta{Name: "x1"}}
	o2 := model.KubeResource{TypeMeta: model.TypeMeta{"y"}, ResourceMeta: model.ResourceMeta{Name: "y1"}}
	c.updateKubeObject(s, o1)
	c.updateKubeObject(s, o2)

	c.deleteKubeObjects(s, "y")
	assert.Equal(t, []model.KubeResource{o1}, c.resources[s])
}

func TestUpdateKubeObject(t *testing.T) {
	c := NewWatchCache()
	s := "s"

	expected := []model.KubeResource{
		{TypeMeta: model.TypeMeta{"x"}, ResourceMeta: model.ResourceMeta{Name: "x1"}},
		{TypeMeta: model.TypeMeta{"y"}, ResourceMeta: model.ResourceMeta{Name: "x1"}},
		{TypeMeta: model.TypeMeta{"y"}, ResourceMeta: model.ResourceMeta{Name: "x1", Namespace: "ns2"}},
	}

	for i := range expected {
		c.updateKubeObject(s, expected[i])
	}

	assert.Equal(t, expected, c.resources[s])
}
