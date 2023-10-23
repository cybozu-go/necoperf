package cmd

import (
	"log/slog"
	"os"

	"github.com/cybozu-go/necoperf/internal/constants"
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
		Use:   "start",
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
	cmd.Flags().IntVar(&port, "port", constants.NecoPerfGrpcServerPort, "Port number on which the grpc server runs")
	cmd.Flags().StringVar(&runtimeEndpoint, "runtime-endpoint", "unix:///run/containerd/containerd.sock", "Container runtime endpoint to connect to")
	cmd.Flags().StringVar(&workDir, "work-dir", "/var/necoperf", "Directory for storing profiling result")

	return cmd
}
