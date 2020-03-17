package service

import (
	"autocli/model"
	"sync"
)

type WatchCache struct {
	Resources map[string][]model.KubeResource
	mu        *sync.RWMutex
}

func (c *WatchCache) DeleteKubeObjects(s string, kind string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.Resources[s]
	if !ok {
		return
	}

	newObjects := []model.KubeResource{}
	for i := range os {
		if os[i].Kind != kind {
			newObjects = append(newObjects, os[i])
		}
	}

	c.Resources[s] = newObjects
}

func (c *WatchCache) DeleteKubeObject(s string, o model.KubeResource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.Resources[s]
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
		c.Resources[s] = os
	}
}

func (c *WatchCache) UpdateKubeObject(s string, o model.KubeResource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	os, ok := c.Resources[s]
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
	c.Resources[s] = os
}

func NewWatchCache() *WatchCache {
	c := &WatchCache{}
	c.mu = &sync.RWMutex{}
	c.Resources = make(map[string][]model.KubeResource)
	return c
}
