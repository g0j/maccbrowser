package runner

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

type Runner struct {
	cli *client.Client
}

func NewRunner(ctx context.Context) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	return &Runner{
		cli: cli,
	}, nil
}

func (r *Runner) Close() error {
	if err := r.cli.Close(); err != nil {
		return fmt.Errorf("close client: %w", err)
	}

	return nil
}
