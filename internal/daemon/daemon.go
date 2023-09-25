package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/cybozu-go/necoperf/internal/resource"
	"github.com/cybozu-go/necoperf/internal/rpc"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/oklog/run"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
)

type DaemonServer struct {
	logger   *slog.Logger
	server   *grpc.Server
	port     int
	endpoint string
	workDir  string
	rpc.UnimplementedNecoPerfServer
	container    *resource.Container
	perfExecuter *resource.PerfExecuter
}

const (
	minTime    = 30 * time.Second
	criTimeout = 30 * time.Second
)

func New(logger *slog.Logger, port int, endpoint, workDir string) (*DaemonServer, error) {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	kep := keepalive.EnforcementPolicy{
		MinTime: minTime,
	}

	serv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(logger), opts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(InterceptorLogger(logger), opts...),
		),
		grpc.KeepaliveEnforcementPolicy(
			kep,
		),
	)

	return &DaemonServer{
		logger:   logger,
		server:   serv,
		port:     port,
		endpoint: endpoint,
		workDir:  workDir,
	}, nil
}

// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/examples/slog/example_test.go
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), "msg", msg, fields)
	})
}

func (d *DaemonServer) Start() error {
	rpc.RegisterNecoPerfServer(d.server, d)
	hs := health.NewServer()
	healthpb.RegisterHealthServer(d.server, hs)
	hs.Resume()
	reflection.Register(d.server)

	if err := d.setupWorkDir(); err != nil {
		return err
	}

	if err := d.setupContainer(); err != nil {
		return err
	}

	if err := d.setupPerfExecuter(); err != nil {
		return err
	}

	g := &run.Group{}
	g.Add(func() error {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", d.port))
		if err != nil {
			return err
		}
		d.logger.Info("gRPC server is running", "port", d.port)
		defer l.Close()

		return d.server.Serve(l)
	}, func(err error) {
		d.logger.Error("gRPC server shutdown", "error", err)
		d.server.GracefulStop()
		d.server.Stop()
	})

	return g.Run()
}

func (d *DaemonServer) setupWorkDir() error {
	_, err := os.Stat(d.workDir)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(d.workDir, 0755); err != nil {
		return err
	}
	return nil
}

func (d *DaemonServer) setupContainer() error {
	client, err := remote.NewRemoteRuntimeService(d.endpoint, criTimeout, nil)
	if err != nil {
		return err
	}

	container := resource.NewContainer(d.logger, client)
	d.container = container
	return nil
}

func (d *DaemonServer) setupPerfExecuter() error {
	perfExecuter, err := resource.NewPerfExecuter(d.logger)
	if err != nil {
		return err
	}
	d.perfExecuter = perfExecuter
	return nil
}
