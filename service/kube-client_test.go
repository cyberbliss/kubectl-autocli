package service

import (
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

func createTestNodes(client kubernetes.Interface, names ...string) {
	for _, name := range names {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		client.CoreV1().Nodes().Create(node)
	}
}
