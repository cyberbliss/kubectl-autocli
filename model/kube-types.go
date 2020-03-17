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
