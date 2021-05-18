package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const imageName = "selenoid/vnc"

type Runner struct {
	ctx context.Context
	cli *client.Client

	containersMx sync.Mutex
	containers   map[string]string
}

type RunOptions struct {
	GUID          string
	ChromeVersion string
}

func NewRunner(ctx context.Context) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	return &Runner{
		ctx:        ctx,
		cli:        cli,
		containers: make(map[string]string),
	}, nil
}

func (r *Runner) Close() error {
	timeout := time.Second * 5

	r.containersMx.Lock()
	defer r.containersMx.Unlock()

	for _, c := range r.containers {
		err := r.cli.ContainerStop(context.Background(), c, &timeout)
		if err != nil {
			return fmt.Errorf("stop container %s: %w", c, err)
		}

		err = r.cli.ContainerRemove(context.Background(), c, types.ContainerRemoveOptions{})
		if err != nil {
			return fmt.Errorf("remove container %s: %w", c, err)
		}
	}

	if err := r.cli.Close(); err != nil {
		return fmt.Errorf("close client: %w", err)
	}

	return nil
}

func (r *Runner) CheckImage(tagname string) (bool, error) {
	ctx, cancel := context.WithTimeout(r.ctx, time.Second)
	defer cancel()

	images, err := r.cli.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: fmt.Sprintf("%s:%s", imageName, tagname),
		}),
	})
	if err != nil {
		return false, fmt.Errorf("get image list: %w", err)
	}

	return len(images) > 0, nil
}

func (r *Runner) PullImage(tagname string) error {
	ctx, cancel := context.WithTimeout(r.ctx, time.Minute)
	defer cancel()

	image, err := r.cli.ImagePull(ctx, fmt.Sprintf("%s:%s", imageName, tagname), types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("image pull: %w", err)
	}

	_, err = io.Copy(os.Stdout, image)
	if err != nil {
		return fmt.Errorf("output status: %w", err)
	}

	err = image.Close()
	if err != nil {
		return fmt.Errorf("closing reader: %w", err)
	}

	return nil
}

func (r *Runner) RunContainer(tagname, guid string) error {
	ctx, cancel := context.WithTimeout(r.ctx, time.Minute)
	defer cancel()

	resp, err := r.cli.ContainerCreate(ctx,
		&container.Config{
			Image: fmt.Sprintf("%s:%s", imageName, tagname),
		},
		&container.HostConfig{},
		&network.NetworkingConfig{},
		nil,
		guid)
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	r.containersMx.Lock()
	r.containers[guid] = resp.ID
	r.containersMx.Unlock()

	if err := r.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("starting container: %w", err)
	}

	return nil
}

func (r *Runner) RunBrowser(opts RunOptions) error {
	tag := fmt.Sprintf("chrome_%s", opts.ChromeVersion)

	ok, err := r.CheckImage(tag)
	if err != nil {
		return err
	}

	if !ok {
		err = r.PullImage(tag)
		if err != nil {
			return err
		}
	}

	err = r.RunContainer(tag, opts.GUID)
	if err != nil {
		return err
	}

	return nil
}
