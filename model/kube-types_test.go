package model

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestByKindNSNameSort(t *testing.T) {
	expected := make([]KubeResource, 0)
	for _, kind := range []string{"deployment", "pod"} {
		for _, ns := range []string{"abc", "def"} {
			for _, name := range []string{"a","b","c"} {
				r := KubeResource{
					TypeMeta:     TypeMeta{Kind: kind},
					ResourceMeta: ResourceMeta{
						Name: name,
						Namespace: ns,
					},
				}
				expected = append(expected, r)
			}
		}
	}
	actual := expected

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(actual), func(i, j int) { actual[i], actual[j] = actual[j], actual[i] })

	sort.Sort(ByKindNSName(actual))

	assert.Equal(t, expected, actual)
}
