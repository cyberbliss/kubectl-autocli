package main

import (
	"autocli/cmd"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use:          "kubectl-ac",
	Short:        "kubectl-ac",
	SilenceUsage: true,
}

func main() {
	b := cmd.NewBuilder()
	//rootCmd := cmd.NewGetCommand(b)
	RootCmd.AddCommand(cmd.NewVersionCommand(b))
	RootCmd.AddCommand(cmd.NewWatchCommand(b))
	RootCmd.AddCommand(cmd.NewResourcesCommand(b))
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
