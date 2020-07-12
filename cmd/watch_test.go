package cmd

import (
	"autocli/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRunWatch(t *testing.T) {
	b := NewTestBuilder()
	servers := []string{"prod", "dev"}
	cmd := NewWatchCommand(b)
	cmd.Flags().Set("kubeconfig", "test_data/kubeconfig_valid")
	cmd.Flags().Set("port", "33044")
	//cmd.Flags().Set("verbose", "true")

	go cmd.RunE(cmd, servers)
	time.Sleep(50 * time.Millisecond)

	bind, err := GetBind(cmd)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	client, err := b.WatchClient(bind, "", "", "")
	if err != nil {
		t.Errorf("could not create client to autocli: %s", err)
	}
	wf := makeFilter("prod", "", "pod")
	var kr []model.KubeResource
	for {
		kr, err = client.Resources(wf)
		if err != nil {
			t.Error(err)
			break
		}
		if len(kr) > 0 {
			break
		}
	}

	assert.Equal(t, "pod", kr[0].Kind)
	assert.Equal(t, "ns1-pod", kr[0].Name)
	assert.Equal(t, "ns1", kr[0].Namespace)

	wf = makeFilter("dev", "ns2", "pod")
	for {
		kr, err = client.Resources(wf)
		if err != nil {
			t.Error(err)
			break
		}
		if len(kr) > 0 {
			break
		}
	}

	assert.Equal(t, "pod", kr[0].Kind)
	assert.Equal(t, "ns2-pod", kr[0].Name)
	assert.Equal(t, "ns2", kr[0].Namespace)

	wf = makeFilter("prod", "", "node")
	kr, err = client.Resources(wf)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(kr))
	assert.Equal(t, "prodnode2", kr[1].Name)
	assert.Equal(t, "NotReady", kr[1].Status)
}
