package daemon

import (
	"io"
	"os"
	"time"

	"github.com/cybozu-go/necoperf/internal/rpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxTimeout = 10 * time.Minute // 10 minute
	weight     = 1
)

func (d *DaemonServer) Profile(req *rpc.PerfProfileRequest, stream rpc.NecoPerf_ProfileServer) error {
	eg, ctx := errgroup.WithContext(stream.Context())
	containerID := req.GetContainerId()
	if len(containerID) == 0 {
		err := status.Error(codes.InvalidArgument, "container ID is not set")
		return err
	}

	timeoutpb := req.GetTimeout()
	if !timeoutpb.IsValid() {
		err := status.Errorf(codes.InvalidArgument, "timeout is invalid value")
		return err
	}

	timeout := timeoutpb.AsDuration()
	if timeout > maxTimeout {
		return status.Errorf(codes.InvalidArgument, "timeout is too long %q", timeout)
	}

	pid, err := d.container.GetPidFromContainerID(ctx, containerID)
	if err != nil {
		return err
	}
	if pid < 1 {
		err := status.Error(codes.Internal, "invalid PID is returned from CRI API")
		return err
	}

	err = d.semaphore.Acquire(ctx, weight)
	if err != nil {
		return err
	}

	var scriptDataPath string
	defer os.Remove(scriptDataPath)

	eg.Go(func() error {
		defer d.semaphore.Release(weight)

		profileDataPath, err := d.perfExecuter.ExecRecord(ctx, d.workDir, pid, timeout)
		defer os.Remove(profileDataPath)
		if err != nil {
			return err
		}

		scriptDataPath, err = d.perfExecuter.ExecScript(ctx, profileDataPath, d.workDir)
		if err != nil {
			return err
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	f, err := os.Open(scriptDataPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&rpc.PerfProfileResponse{
			Data: buf[:n],
		}); err != nil {
			return err
		}
	}

	return nil
}
