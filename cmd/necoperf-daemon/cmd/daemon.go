package cmd

import (
	"log/slog"
	"os"

	"github.com/cybozu-go/necoperf/internal/daemon"
	"github.com/spf13/cobra"
)

var (
	port            int
	runtimeEndpoint string
	workDir         string
)

func NewDaemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Starts the daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			handler := slog.NewTextHandler(os.Stderr, nil)
			logger := slog.New(handler)
			daemon, err := daemon.New(logger, port, runtimeEndpoint, workDir)
			if err != nil {
				return err
			}

			return daemon.Start()
		},
	}
	cmd.Flags().IntVar(&port, "port", 6543, "Set server port number")
	cmd.Flags().StringVar(&runtimeEndpoint, "runtime-endpoint", "unix:///run/containerd/containerd.sock", "Set container runtime endpoint")
	cmd.Flags().StringVar(&workDir, "work-dir", "/var/necoperf", "Set working directory")

	return cmd
}
