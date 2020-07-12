package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	BuildVersion string = "dev"
	BuildTime    string = ""
)

func NewVersionCommand(b Builder) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			RunVersion(b, cmd, args)
		},
	}

	return cmd
}

func RunVersion(b Builder, cmd *cobra.Command, args []string) {
	fmt.Fprintln(b.StdOut(), "kubectl-ac")
	fmt.Fprintf(b.StdOut(), "Version: %s\n", BuildVersion)
	fmt.Fprintf(b.StdOut(),"Build time: %s\n", BuildTime)
}
