package cmd

import (
	"autocli/model"
	"errors"
	"fmt"
	"net/rpc"
	"sort"
	"strings"
	"sync"

	strUtil "github.com/agrison/go-commons-lang/stringUtils"
	log "github.com/sirupsen/logrus"

)

type WatchCache struct {
	resources map[string][]model.KubeResource
	mu        *sync.RWMutex
}

type WatchFilter struct {
	Context   string
	Namespace string
	Kind      string
}

func (c *WatchCache) deleteKubeObjects(s string, kind string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.resources[s]
	if !ok {
		return
	}

	newObjects := []model.KubeResource{}
	for i := range os {
		if os[i].Kind != kind {
			newObjects = append(newObjects, os[i])
		}
	}

	c.resources[s] = newObjects
}

func (c *WatchCache) deleteKubeObject(s string, o model.KubeResource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.resources[s]
	if !ok {
		return
	}

	idx := -1
	for i := range os {
		if os[i].Name == o.Name && os[i].Namespace == o.Namespace {
			idx = i
			break
		}
	}

	if idx >= 0 {
		os = append(os[:idx], os[idx+1:]...)
		c.resources[s] = os
	}
}

func (c *WatchCache) updateKubeObject(s string, o model.KubeResource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.resources[s]
	if !ok {
		os = make([]model.KubeResource, 0)
	}

	found := false
	for i := range os {
		if os[i].Name == o.Name && os[i].Namespace == o.Namespace && os[i].Kind == o.Kind {
			os[i] = o
			found = true
			break
		}
	}

	if !found {
		os = append(os, o)
	}
	c.resources[s] = os
}

func (c *WatchCache) Resources(f *WatchFilter, kr *[]model.KubeResource) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	log.WithField("filter", f).Debug("Received request for resources")

	if f == nil {
		return errors.New("cannot find resources with nil filter")
	}

	keys := []string{}

	for k, _ := range c.resources {
		if f.Context == "" || strings.EqualFold(f.Context, k) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		log.WithField("context", f.Context).Error("unknown context")
		return fmt.Errorf("unknown context %s", f.Context)
	}

	sort.Strings(keys)

	res := []model.KubeResource{}
	for _, k := range keys {
		for _, r := range c.resources[k] {
			if strings.EqualFold(r.Kind, f.Kind) &&
				(f.Namespace == "" || r.Kind == "namespace" || strings.EqualFold(r.Namespace, f.Namespace)) {
				res = append(res, r)
			}
		}
	}

	log.WithField("filter", f).WithField("resources", res).Debug("Returning result for resources")
	*kr = res

	return nil
}

func (c *WatchCache) Status(ctx *string, ct *int) error {
	log.WithField("context", ctx).Debug("Received request for status")
	if strUtil.IsBlank(*ctx) {
		return errors.New("context cannot be blank")
	}

	if res, exists := c.resources[*ctx]; exists {
		*ct = len(res)
		return nil
	} else {
		return fmt.Errorf("kube context %s not found", *ctx)
	}

}

func NewWatchCache() *WatchCache {
	c := &WatchCache{}
	c.mu = &sync.RWMutex{}
	c.resources = make(map[string][]model.KubeResource)
	return c
}

type WatchClient interface {
	Resources(f WatchFilter) ([]model.KubeResource, error)
	Status(c string) (int, error)
}

type WatchClientDefault struct {
	conn *rpc.Client
	builderType string
}

func NewWatchClient(address, builderType, rpcPath string) (*WatchClientDefault, error) {
	var connection *rpc.Client
	var err error

	if strUtil.IsBlank(rpcPath) {
		connection, err = rpc.DialHTTP("tcp", address)
	} else {
		connection, err = rpc.DialHTTPPath("tcp", address, rpcPath)
	}
	if err != nil {
		return nil, err
	}

	return &WatchClientDefault{
		conn: connection,
		builderType: builderType,
	}, nil
}

func (wc *WatchClientDefault) Resources(f WatchFilter) ([]model.KubeResource, error) {
	var res []model.KubeResource
	//err := wc.conn.Call("WatchCache.Resources", f, &res)
	sm := wc.builderType + ".Resources"
	err := wc.conn.Call(sm, f, &res)
	return res, err
}

func (wc *WatchClientDefault) Status(c string) (int, error) {
	log.Debug("in status")
	var resourceCount int
	sm := wc.builderType + ".Status"
	err := wc.conn.Call(sm, c, &resourceCount)
	return resourceCount, err
}
