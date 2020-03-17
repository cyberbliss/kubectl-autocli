package cmd

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

func NewGetCommand(b Builder) *cobra.Command {
	var cmd = &cobra.Command{
		Use:                        "get",
		Aliases:                    nil,
		Short:                      "get pods",
		Long:                       "get pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGet(b, cmd, args)
		},
	}

	return cmd
}

func RunGet(b Builder, cmd *cobra.Command, args []string) error {

	in := prompt.Input(">>> ", b.PodCompleter,

		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray))
	//fmt.Println("Your input: " + in)
	fmt.Fprintf(b.StdOut(), "Your input: %s\n", in)

	return nil
}

