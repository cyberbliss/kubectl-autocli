package service

import (
	"autocli/model"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"strings"

	log "github.com/sirupsen/logrus"
)

type KubeClient interface {
	Ping(context string) error
	WatchResources(context, kind string, out chan *model.ResourceEvent) error
}

type DefaultKubeClient struct {
	clients map[string]kubernetes.Interface
}

func (d *DefaultKubeClient) Ping(context string) error {
	client, ok := d.clients[context]
	if !ok {
		return fmt.Errorf("Context not found: %s", context)
	}
	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return err
	}
	log.Debugf("Nodes: %s", nodes)
	if len(nodes.Items) < 1 {
		return fmt.Errorf("no nodes available for context: %s", context)
	}

	return nil
}

func (d *DefaultKubeClient) WatchResources(context, kind string, out chan *model.ResourceEvent) error{
	client, ok := d.clients[context]
	if !ok {
		return fmt.Errorf("context not found: %s", context)
	}

	switch kind {
	case "pod":
		ch, err := d.watchPods(client, v1.NamespaceAll)
		if err != nil {
			return fmt.Errorf("watching pods failed: %s", err)
		}
		for event := range ch {
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				return fmt.Errorf("unexpected type: %s", pod)
			}
			var evt model.ResourceEvent
			log.Debug(event.Type)
			switch event.Type {
			case watch.Added:
				evt.Type = model.Added
			case watch.Deleted:
				evt.Type = model.Deleted
			case watch.Modified:
				evt.Type = model.Modified
			}
			var status string
			if strings.TrimSpace(pod.Status.Message) == "" {
				status = string(pod.Status.Phase)
			} else {
				status = fmt.Sprintf("%s: %s",pod.Status.Phase, pod.Status.Message)
			}
			evt.Resource = &model.KubeResource{
				TypeMeta:   model.TypeMeta{Kind: "pod"},
				ResourceMeta: model.ResourceMeta{
					Name: pod.Name,
					Namespace: pod.Namespace,
					ResourceVersion: pod.ResourceVersion,
					Status: status,
				},
			}
			out <- &evt
		}
	default:
		return fmt.Errorf("unsupported kind: %s", kind)
	}

	return nil
}

func (d *DefaultKubeClient) watchPods(client kubernetes.Interface, ns string) (out <-chan watch.Event, err error) {
	req, err := client.CoreV1().Pods(ns).Watch(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return req.ResultChan(), nil
}

func NewKubeClient(clients map[string]kubernetes.Interface) KubeClient {
	dkc := &DefaultKubeClient{
		clients: clients,
	}

	return dkc
}
