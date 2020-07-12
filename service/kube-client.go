package service

import (
	"autocli/model"
	"context"
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
	GetResources(context, kind string) ([]model.KubeResource, error)
}

type DefaultKubeClient struct {
	clients map[string]kubernetes.Interface
}

func (d *DefaultKubeClient) Ping(ctx string) error {
	client, ok := d.clients[ctx]
	if !ok {
		return fmt.Errorf("context not found: %s", ctx)
	}
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return err
	}
	log.Debugf("Nodes: %s", nodes)
	if len(nodes.Items) < 1 {
		return fmt.Errorf("no nodes available for context: %s", ctx)
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
			var cNames []model.ContainerMeta
			if strings.TrimSpace(pod.Status.Message) == "" {
				status = string(pod.Status.Phase)
			} else {
				status = fmt.Sprintf("%s: %s",pod.Status.Phase, pod.Status.Message)
			}

			//get all the container names (including init ones if there are any)
			if len(pod.Spec.InitContainers) > 0 {
				for _,c := range pod.Spec.InitContainers {
					cNames = append(cNames, model.ContainerMeta{
						Name: c.Name,
						Type: "Init Container",
					})
				}
			}
			for _,c := range pod.Spec.Containers {
				cNames = append(cNames, model.ContainerMeta{
					Name: c.Name,
					Type: "Container",
				})
			}

			evt.Resource = &model.KubeResource{
				TypeMeta:   model.TypeMeta{Kind: "pod"},
				ResourceMeta: model.ResourceMeta{
					Name: pod.Name,
					Namespace: pod.Namespace,
					ResourceVersion: pod.ResourceVersion,
					Status: status,
					ContainerNames: cNames,
				},
			}
			out <- &evt
		}
	default:
		return fmt.Errorf("unsupported kind: %s", kind)
	}

	return nil
}

func (d *DefaultKubeClient) GetResources(ctx, kind string) ([]model.KubeResource, error) {
	var resources []model.KubeResource

	client, ok := d.clients[ctx]
	if !ok {
		return []model.KubeResource{}, fmt.Errorf("context not found: %s", ctx)
	}

	switch kind {
	case "node":
		nodes, err := d.getNodes(client)
		if err != nil {
			return []model.KubeResource{}, err
		}
		log.Debug(nodes)
		for _, node := range nodes.Items {
			AddToKubeResources(&resources, "node", node.Name, node.Namespace, node.ResourceVersion, determineNodeStatus(node.Status.Conditions))
		}
		return resources, nil
	default:
		return []model.KubeResource{}, fmt.Errorf("unsupported kind: %s", kind)
	}
}

func (d *DefaultKubeClient) watchPods(client kubernetes.Interface, ns string) (out <-chan watch.Event, err error) {
	req, err := client.CoreV1().Pods(ns).Watch(context.TODO(),metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return req.ResultChan(), nil
}

func (d *DefaultKubeClient) getNodes(client kubernetes.Interface) (*v1.NodeList, error) {
	return client.CoreV1().Nodes().List(context.TODO(),metav1.ListOptions{})
}

func (d *DefaultKubeClient) getConfigMaps(client kubernetes.Interface, ns string) (*v1.ConfigMapList, error) {
	return client.CoreV1().ConfigMaps(ns).List(context.TODO(),metav1.ListOptions{})
}

func (d *DefaultKubeClient) getServices(client kubernetes.Interface, ns string) (*v1.ServiceList, error) {
	return client.CoreV1().Services(ns).List(context.TODO(),metav1.ListOptions{})
}

func NewKubeClient(clients map[string]kubernetes.Interface) KubeClient {
	dkc := &DefaultKubeClient{
		clients: clients,
	}

	return dkc
}

func AddToKubeResources(resourcesPtr *[]model.KubeResource, tm, name, ns, rv, status string) {
	//NOTE: need to pass the KubeObject as a pointer because I'm altering the actual slice by appending to it
	res := *resourcesPtr
	t := model.TypeMeta{Kind: tm}
	*resourcesPtr = append(res, model.KubeResource{
		TypeMeta:     t,
		ResourceMeta: model.ResourceMeta{
			Name:            name,
			Namespace:       ns,
			ResourceVersion: rv,
			Status:          status,
			ContainerNames:  nil,
		},
	})
}

func determineNodeStatus(conditions []v1.NodeCondition) string {
	var status = "Unknown"
	for _, v := range conditions {
		if v.Type == v1.NodeReady {
			if v.Status == v1.ConditionTrue {
				status = "Ready"
				break
			}
		}
	}

	return status
}


