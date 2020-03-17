package service

import "github.com/c-bata/go-prompt"

func GetPods() []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "1: pod1", Description: "namespace1"},
		{Text: "2: pod2", Description: "namespace1"},
		{Text: "3: pod1", Description: "namespace2"},
		{Text: "4: pod3", Description: "namespace2"},
	}

	return s
}
