package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

type SuggestList struct {
	Suggestions []SuggestType `json:"suggestions"`
}

type SuggestType struct {
	Text        string `json:"text"`
	Description string `json:"desc"`
}

func completer(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
		{Text: "groups", Description: "Combine users with specific rules"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func fileCompleter(in prompt.Document) []prompt.Suggest {
	jsonFile, err := os.Open("suggestions.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var sl SuggestList
	var s []prompt.Suggest
	json.Unmarshal(byteValue, &sl)
	for i := 0; i < len(sl.Suggestions); i++ {
		s = append(s, prompt.Suggest{
			Text:        fmt.Sprintf("%d: %s", i, sl.Suggestions[i].Text),
			Description: sl.Suggestions[i].Description,
		})
	}

	return prompt.FilterContains(s, in.GetWordBeforeCursor(), true)
}

func executor(t string) {
	//fmt.Println(t)
	cmd := exec.Command("kubectl", "get", "pods", "-n", t)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed with %s\n", err)
	}
}

func main() {
	in := prompt.Input(">>> ", fileCompleter,

		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray))
	fmt.Println("Your input: " + in)
	ns := strings.Split(in, ": ")
	executor(ns[1])
	//p := prompt.New(
	//	executor,
	//	fileCompleter,
	//	prompt.OptionTitle("sql-prompt"),
	//	prompt.OptionHistory([]string{"SELECT * FROM users;"}),
	//	prompt.OptionPrefixTextColor(prompt.Yellow),
	//	prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
	//	prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
	//	prompt.OptionSuggestionBGColor(prompt.DarkGray),
	//	)
	//p.Run()
}
