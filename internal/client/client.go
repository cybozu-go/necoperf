package client

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cybozu-go/necoperf/internal/resource"
	"github.com/cybozu-go/necoperf/internal/rpc"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/durationpb"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	logger    *slog.Logger
	client    rpc.NecoPerfClient
	Discovery *resource.Discovery
	Timeout   time.Duration
}

// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/examples/slog/example_test.go
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), "msg", msg, fields)
	})
}

func New(logger *slog.Logger, timeout time.Duration) (*Client, error) {
	return &Client{
		logger:  logger,
		Timeout: timeout,
	}, nil
}

func (c *Client) Save(dataDir, podName string) (*os.File, error) {
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, err
	}

	profileFilePath := filepath.Join(dataDir + "/" + podName + ".script")
	f, err := os.Create(profileFilePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c *Client) Profile(ctx context.Context, podName, containerID, DataDir string) error {
	t := durationpb.New(c.Timeout)
	req := &rpc.PerfProfileRequest{
		ContainerId: containerID,
		Timeout:     t,
	}

	stream, err := c.client.Profile(ctx, req)
	if err != nil {
		return err
	}

	f, err := c.Save(DataDir, podName)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(f.Name())
			return err
		}

		_, err = f.Write(resp.Data)
		if err != nil {
			os.Remove(f.Name())
			return err
		}
	}

	return nil
}

func (c *Client) SetupDiscovery() error {
	config, err := config.GetConfig()
	if err != nil {
		return err
	}

	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}

	d, err := resource.NewDiscovery(c.logger, k8sClient)
	if err != nil {
		return err
	}
	c.Discovery = d

	return nil
}

func (c *Client) SetupGrpcClient(addr string) error {
	kp := keepalive.ClientParameters{
		Time: c.Timeout * 3,
	}
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	conn, err := grpc.Dial(
		addr,
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(InterceptorLogger(c.logger), opts...),
		),
		grpc.WithChainStreamInterceptor(
			logging.StreamClientInterceptor(InterceptorLogger(c.logger), opts...),
		),
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithKeepaliveParams(
			kp,
		),
	)
	if err != nil {
		return err
	}
	c.client = rpc.NewNecoPerfClient(conn)

	return nil
}
