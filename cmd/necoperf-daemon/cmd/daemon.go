package cmd

import (
	"log/slog"
	"os"

	"github.com/cybozu-go/necoperf/internal/daemon"
	"github.com/spf13/cobra"
)

var port int
var runtimeEndpoint string

func init() {
	rootCmd.AddCommand(daemonCmd)
	flags := daemonCmd.Flags()
	flags.IntVar(&port, "port", 6543, "Set server port number")
	flags.StringVar(&runtimeEndpoint, "runtime-endpoint", "unix:///run/containerd/containerd.sock", "Set container runtime endpoint")
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		handler := slog.NewTextHandler(os.Stderr, nil)
		logger := slog.New(handler)
		daemon, err := daemon.New(logger, port, runtimeEndpoint, os.TempDir())
		if err != nil {
			return err
		}

		return daemon.Start()
	},
}
