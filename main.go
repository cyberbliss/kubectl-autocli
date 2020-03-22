package main

import (
	"autocli/cmd"
	"os"
)

func main() {
	b := cmd.NewBuilder()
	rootCmd := cmd.NewGetCommand(b)
	rootCmd.AddCommand(cmd.NewVersionCommand(b))
	rootCmd.AddCommand(cmd.NewWatchCommand(b))
	rootCmd.AddCommand(cmd.NewPodsCommand(b))
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
