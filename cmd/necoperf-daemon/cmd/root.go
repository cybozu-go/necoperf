package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "necoperf-daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
