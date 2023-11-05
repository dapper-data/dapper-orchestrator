package orchestrator

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const (
	statusOK = 0

	imageKey = "image"
	envKey   = "env"
)

// ContainerImageMissingErr is returned when the ExecutionContext passed to
// NewContainerProcess doesn't contain tke key "image"
//
// To fix this, ensure that a container image is set
type ContainerImageMissingErr struct{}

// Error implements the error interface and returns a contextual message
//
// This error, while simple and (at least on the face of it) an over-engineered
// version of fmt.Errorf("container image missing"), is verbosely implemented
// so that callers may use errors.Is(err, orchestrator.ContainerImageMissingErr)
// to handle error cases better
func (e ContainerImageMissingErr) Error() string {
	return "container image missing"
}

// ContainerNonZeroExit is returned when the container exists with anything other
// than exit code 0
//
// Container logs should shed light on what went wrong
type ContainerNonZeroExit int64

// Error returns the error message associated with this error
func (e ContainerNonZeroExit) Error() string {
	return fmt.Sprintf("process exited with code %d", int64(e))
}

// ContainerProcess allows for processes to be run via a container
type ContainerProcess struct {
	image         string
	additionalEnv []string
	c             *client.Client

	config ProcessConfig
}

// NewContainerProcess connects to a container socket, and returns a
// ContainerProcess which can be then used to run jobs
func NewContainerProcess(conf ProcessConfig) (c ContainerProcess, err error) {
	var ok bool

	c.config = conf
	c.image, ok = conf.ExecutionContext[imageKey]
	if !ok {
		err = ContainerImageMissingErr{}

		return
	}

	c.additionalEnv = strings.Split(conf.ExecutionContext[envKey], ",")

	c.c, err = client.NewClientWithOpts(client.FromEnv)

	return
}

// ID returns a unique ID for a process manager
func (c ContainerProcess) ID() string {
	return c.config.ID()
}

// Run takes an Event, and passes it to a container to run
func (c ContainerProcess) Run(ctx context.Context, e Event) (ps ProcessStatus, err error) {
	name := c.deriveName()

	ps = ProcessStatus{
		Name:   name,
		Status: ProcessUnstarted,
		Logs:   make([]string, 0),
	}

	cont, err := c.c.ContainerCreate(
		ctx,
		&container.Config{
			Image:        c.image,
			Env:          c.env(e),
			AttachStdout: false,
			AttachStderr: true,
		},
		&container.HostConfig{
			NetworkMode: container.NetworkMode("host"),
		}, nil, nil, name)
	if err != nil {
		return
	}

	err = c.c.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return
	}

	ps.Status = ProcessUnknown

	wrC, errC := c.c.ContainerWait(ctx, cont.ID, "")
	select {
	case err = <-errC:
		ps.Status = ProcessFail

		return

	case wr := <-wrC:
		rc, err := c.c.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{
			ShowStdout: false,
			ShowStderr: true,
		})
		if err != nil {
			return ps, err
		}

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		_, err = stdcopy.StdCopy(stdout, stderr, rc)
		if err != nil {
			return ps, err
		}

		ps.Logs = strings.Split(stderr.String(), "\n")

		switch wr.StatusCode {
		case statusOK:
			ps.Status = ProcessSuccess

		default:
			ps.Status = ProcessFail

			return ps, ContainerNonZeroExit(wr.StatusCode)
		}
	}

	return
}

func (c ContainerProcess) deriveName() string {
	return fmt.Sprintf("%s_%v", c.ID(), time.Now().UnixMicro())
}

func (c ContainerProcess) env(e Event) (out []string) {
	ev, err := e.JSON()
	if err != nil {
		return c.additionalEnv
	}

	out = []string{
		fmt.Sprintf("PIPELINE_EVENT=%q", base64.StdEncoding.EncodeToString([]byte(ev))),
	}

	return append(out, c.additionalEnv...)
}
