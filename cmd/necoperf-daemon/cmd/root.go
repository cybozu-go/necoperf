package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "necoperf-daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	return cmd
}

func Execute() {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewDaemonCommand())
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
