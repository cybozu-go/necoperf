package cmd

import (
	"log/slog"
	"os"

	"github.com/cybozu-go/necoperf/internal/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	flags := daemonCmd.Flags()
	flags.IntP("port", "p", 6543, "Set server port number")
	flags.StringP("runtime-endpoint", "r", "unix:///run/containerd/containerd.sock", "Set container runtime endpoint")
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		endpoint, err := cmd.Flags().GetString("runtime-endpoint")
		if err != nil {
			return err
		}

		handler := slog.NewTextHandler(os.Stderr, nil)
		logger := slog.New(handler)
		daemon, err := daemon.New(logger, port, endpoint, os.TempDir())
		if err != nil {
			return err
		}
		return daemon.Start()
	},
}
