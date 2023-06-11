package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/loft-sh/devpod/pkg/client"
)

type Hetzner struct {
	client *hcloud.Client
}

func NewHetzner(token string) *Hetzner {
	return &Hetzner{
		client: hcloud.NewClient(hcloud.WithToken(token)),
	}
}

func (h *Hetzner) Create(ctx context.Context, req interface{}, diskSize int) error {
	return nil
}

func (h *Hetzner) Delete(ctx context.Context, name string) error {
	return nil
}

func (h *Hetzner) GetByName(ctx context.Context, name string) (interface{}, error) {
	return nil, nil
}

func (h *Hetzner) Init(ctx context.Context) error {
	return nil
}

func (h *Hetzner) Status(ctx context.Context, name string) (client.Status, error) {
	return client.Status("@todo"), nil
}

func (h *Hetzner) Stop(ctx context.Context, name string) error {
	return nil
}
