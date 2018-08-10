package main

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var timeout = EXEC_TIMEOUT

var defaultcfg = container.Config{
	Image:           "godwhoa/sandbox:latest",
	AttachStderr:    true,
	AttachStdout:    true,
	NetworkDisabled: true,
	StopTimeout:     &timeout,
}

var defaulthostcfg = container.HostConfig{
	Binds:       []string{"main.c:/src/main.c:ro"},
	CapDrop:     strslice.StrSlice{"ALL"},
	SecurityOpt: []string{"no-new-privileges", seccomp_opt},
	Resources: container.Resources{
		CPUPeriod: 25000, // both somehow limit cpu to 25%
		CPUQuota:  6250,  // TODO: figure out a better way
		PidsLimit: MAX_PID,
		Memory:    MAX_MEM,
	},
	AutoRemove: false,
}

func rmContainer(ctx context.Context, c *client.Client) func(string) {
	return func(cid string) {
		c.ContainerKill(context.Background(), cid, "SIGKILL")
		c.ContainerRemove(ctx, cid, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		log.Printf("contained=%s removed", cid)
	}
}

func runContainer(ctx context.Context, c *client.Client, containercfg container.Config, hostcfg container.HostConfig) (string, io.ReadCloser, error) {
	rm := rmContainer(ctx, c)

	created, err := c.ContainerCreate(ctx,
		&containercfg,
		&hostcfg,
		nil,
		"",
	)
	if err != nil {
		return "", nil, err
	}

	if err := c.ContainerStart(ctx, created.ID, types.ContainerStartOptions{}); err != nil {
		rm(created.ID)
		return "", nil, err
	}

	logstream, err := c.ContainerLogs(ctx, created.ID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		rm(created.ID)
		return "", nil, err
	}
	log.Printf("container=%s created", created.ID)

	return created.ID, logstream, nil

}

// Takes a duplexed stdout/err stream and spits out stdout/err []byte
func splitOutputs(logstream io.ReadCloser) ([]byte, []byte) {
	var stdout, stderr bytes.Buffer
	stdcopy.StdCopy(&stdout, &stderr, logstream)
	logstream.Close()
	return stdout.Bytes(), stderr.Bytes()
}

func NewDockerRunner() *DockerRunner {
	c, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return &DockerRunner{client: c}
}

// DockerRunner implementes Runner interface and uses docker under the hood
type DockerRunner struct {
	client *client.Client
}

// Run runs source file inside a docker container and returns stdout/err
func (runner *DockerRunner) Run(ctx context.Context, srcfile string) ([]byte, []byte, error) {
	hostcfg := defaulthostcfg
	hostcfg.Binds = []string{srcfile + ":/src/main.c:ro"}

	cid, logstream, err := runContainer(ctx, runner.client, defaultcfg, hostcfg)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	stdout, stderr := splitOutputs(logstream)

	rmContainer(ctx, runner.client)(cid)

	return stdout, stderr, nil
}
