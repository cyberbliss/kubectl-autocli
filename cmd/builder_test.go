package cmd

import (
	"autocli/model"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/assert"
)

func TestPodCompleter(t *testing.T) {
	b := &DefaultBuilder{}
	resources := make([]model.KubeResource, 0)
	expected := make([]prompt.Suggest, 0)
	for _, ns := range []string{"ns1", "ns2", "ns3"} {
		for _, kind := range []string{"pod"} {
			for _, name := range []string{"a", "b", "c"} {

				r := model.KubeResource{
					model.TypeMeta{Kind: kind},
					model.ResourceMeta{Name: name, Namespace: ns},
				}
				resources = append(resources, r)
				e := prompt.Suggest{
					Text:        name,
					Description: ns,
				}
				expected = append(expected, e)
			}
		}
	}

	b.PopulateSuggestions(&resources)

	in := prompt.Document{}
	actual := b.PodCompleter(in)
	assert.Equal(t, expected, actual)
}
