package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "necoperf-cli",
		Short: "necoperf-cli is a command line tool for necoperf",
		Long:  "necoperf-cli is a command line tool for necoperf",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	return cmd
}

func Execute() {
	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewProfileCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
