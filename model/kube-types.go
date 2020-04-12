package model

type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
)

type ResourceMeta struct {
	Name            string
	Namespace       string
	ResourceVersion string
	Status			string
}

type TypeMeta struct {
	Kind string
}

type KubeResource struct {
	TypeMeta
	ResourceMeta
}

type ResourceEvent struct {
	Type   EventType
	Resource *KubeResource
}

//******* Sorting functions *******

//Sort by kind, namespace, name
type ByKindNSName []KubeResource

func (kresources ByKindNSName) Len() int {
	return len(kresources)
}

func (kresources ByKindNSName) Swap(i, j int) {
	kresources[i], kresources[j] = kresources[j], kresources[i]
}

func (kresources ByKindNSName) Less(i, j int) bool {
	if kresources[i].Kind < kresources[j].Kind {
		return true
	}
	if kresources[i].Kind > kresources[j].Kind {
		return false
	}
	if kresources[i].Namespace < kresources[j].Namespace {
		return true
	}
	if kresources[i].Namespace > kresources[j].Namespace {
		return false
	}
	return kresources[i].Name < kresources[j].Name
}

