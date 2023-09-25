package daemon

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/cybozu-go/necoperf/internal/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/durationpb"
)

func server(ctx context.Context) (rpc.NecoPerfClient, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	serv := grpc.NewServer()
	rpc.RegisterNecoPerfServer(serv, &DaemonServer{})
	go func() {
		if err := serv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				return lis.Dial()
			},
		),
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Fatal(err)
		}
		serv.Stop()
	}

	client := rpc.NewNecoPerfClient(conn)
	return client, closer
}

const (
	tooLongTimeout = 10 * time.Hour
	timeout        = 1 * time.Second
	containerID    = "test-container"
)

func TestProfile(t *testing.T) {
	ctx := context.Background()
	client, closer := server(ctx)
	defer closer()

	type expected struct {
		out *rpc.PerfProfileResponse
		err error
	}

	tests := map[string]struct {
		in       *rpc.PerfProfileRequest
		expected expected
	}{
		"notSetTimeout": {
			in: &rpc.PerfProfileRequest{
				ContainerId: containerID,
			},
			expected: expected{
				out: nil,
				err: fmt.Errorf("rpc error: code = InvalidArgument desc = timeout is invalid value"),
			},
		},
		"tooLongTimeout": {
			in: &rpc.PerfProfileRequest{
				ContainerId: containerID,
				Timeout:     durationpb.New(tooLongTimeout),
			},
			expected: expected{
				out: nil,
				err: fmt.Errorf("rpc error: code = InvalidArgument desc = timeout is too long %q", tooLongTimeout),
			},
		},
		"notSetContainerID": {
			in: &rpc.PerfProfileRequest{
				Timeout: durationpb.New(timeout),
			},
			expected: expected{
				out: nil,
				err: fmt.Errorf("rpc error: code = InvalidArgument desc = container ID is not set"),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			stream, err := client.Profile(ctx, tt.in)
			if err != nil {
				t.Fatal(err)
			}

			_, err = stream.Recv()
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Want: %q\nGot: %q", tt.expected.err, err.Error())
				}
			}
		})
	}

}
