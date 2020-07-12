package service

import (
	"autocli/model"
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestPing(t *testing.T) {
	clients := make(map[string]kubernetes.Interface)
	clients["test"] = testclient.NewSimpleClientset()
	kc := NewKubeClient(clients)
	createTestNodes(clients["test"],"test")
	err := kc.Ping("test")
	assert.NoError(t,err)
}

func TestPingError(t *testing.T) {
	clients := make(map[string]kubernetes.Interface)
	kc := NewKubeClient(clients)
	err := kc.Ping("test")
	assert.Error(t, err)
}

func TestGetNodes(t *testing.T) {
	clients := make(map[string]kubernetes.Interface)
	clients["test"] = testclient.NewSimpleClientset()
	kc := NewKubeClient(clients)
	createTestNodes(clients["test"],"test1","test2")
	res, err := kc.GetResources("test", "node")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := []model.KubeResource{
		{
			TypeMeta:     model.TypeMeta{Kind: "node"},
			ResourceMeta: model.ResourceMeta{
				Name:            "test1",
				Namespace:       "",
				ResourceVersion: "",
				Status:          "Ready",
				ContainerNames:  nil,
			},
		},
		{
			TypeMeta:     model.TypeMeta{Kind: "node"},
			ResourceMeta: model.ResourceMeta{
				Name:            "test2",
				Namespace:       "",
				ResourceVersion: "",
				Status:          "Ready",
				ContainerNames:  nil,
			},
		},
	}
	assert.Equal(t, expected, res)
}

func createTestNodes(client kubernetes.Interface, names ...string) {
	for _, name := range names {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionTrue}},
			},
		}
		client.CoreV1().Nodes().Create(context.TODO(),node, metav1.CreateOptions{})
	}
}
