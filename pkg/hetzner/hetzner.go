package hetzner

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"strconv"
	"text/template"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/loft-sh/devpod/pkg/client"
	"github.com/loft-sh/devpod/pkg/ssh"
	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/options"
)

//go:embed cloud-config.yaml
var cloudConfig embed.FS

type Hetzner struct {
	client *hcloud.Client
}

func NewHetzner(token string) *Hetzner {
	return &Hetzner{
		client: hcloud.NewClient(hcloud.WithToken(token)),
	}
}

func (h *Hetzner) BuildServerOptions(ctx context.Context, opts *options.Options) (*hcloud.ServerCreateOpts, *string, error) {
	publicKeyBase, err := ssh.GetPublicKeyBase(opts.MachineFolder)
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return nil, nil, err
	}

	location, _, err := h.client.Location.GetByName(ctx, opts.Region)
	if err != nil {
		return nil, nil, err
	}
	if location == nil {
		return nil, nil, ErrUnknownRegion
	}

	serverType, _, err := h.client.ServerType.GetByName(ctx, opts.MachineType)
	if err != nil {
		return nil, nil, err
	}
	if serverType == nil {
		return nil, nil, ErrUnknownMachineID
	}

	// @todo(sje): work out if DevPod handles different architectures
	image, _, err := h.client.Image.GetByNameAndArchitecture(ctx, opts.DiskImage, hcloud.ArchitectureX86)
	if err != nil {
		return nil, nil, err
	}
	if image == nil {
		return nil, nil, ErrUnknownDiskImage
	}

	return &hcloud.ServerCreateOpts{
		Name:       opts.MachineID,
		Location:   location,
		ServerType: serverType,
		Image:      image,
		Labels: map[string]string{
			"type": "devpod",
		},
	}, hcloud.Ptr[string](string(publicKey)), nil
}

func (h *Hetzner) Create(ctx context.Context, req *hcloud.ServerCreateOpts, diskSize int, publicKey string) error {
	volume, err := h.volumeByName(ctx, req.Name)
	if err != nil {
		return err
	}

	if volume == nil {
		// Create the volume as it doesn't exist
		v, _, err := h.client.Volume.Create(ctx, hcloud.VolumeCreateOpts{
			Location:  req.Location,
			Name:      req.Name,
			Size:      diskSize,
			Format:    hcloud.Ptr[string]("ext4"),
			Automount: hcloud.Ptr[bool](false),
			Labels: map[string]string{
				"type": "devpod",
			},
		})
		if err != nil {
			return err
		}

		volume = v.Volume
	}

	// Generate the config init
	userData, err := generateUserData(req.Name, publicKey, strconv.Itoa(volume.ID))
	if err != nil {
		return err
	}
	// Add to server config
	req.UserData = userData

	// Add volume to the server config
	req.Volumes = []*hcloud.Volume{
		{
			ID: volume.ID,
		},
	}

	// Create the volume
	_, _, err = h.client.Server.Create(ctx, *req)

	return err
}

func (h *Hetzner) Delete(ctx context.Context, name string) error {
	return nil
}

func (h *Hetzner) GetByName(ctx context.Context, name string) (*hcloud.Server, error) {
	servers, _, err := h.client.Server.List(ctx, hcloud.ServerListOpts{Name: name})
	if err != nil {
		return nil, err
	}

	serverLength := len(servers)
	if serverLength > 1 {
		return nil, ErrMultipleServersFound(name)
	}
	if serverLength == 0 {
		return nil, nil
	}

	return servers[0], nil
}

func (h *Hetzner) Init(ctx context.Context) error {
	_, _, err := h.client.Server.List(ctx, hcloud.ServerListOpts{})
	if err != nil {
		return err
	}
	return nil
}

func (h *Hetzner) Status(ctx context.Context, name string) (client.Status, error) {
	server, _, err := h.client.Server.GetByName(ctx, name)
	if err != nil {
		return client.StatusNotFound, err
	}
	if server == nil {
		// No server - check the volume
		volume, err := h.volumeByName(ctx, name)
		if err != nil {
			return client.StatusNotFound, err
		} else if volume != nil {
			return client.StatusStopped, nil
		}

		return client.StatusNotFound, nil
	}

	// Is it busy?
	if server.Status != hcloud.ServerStatusRunning {
		return client.StatusBusy, nil
	}

	// @todo(sje): do we need to check if the cloud-init script is finished? "ssh user@path cloud-init status"

	return client.StatusRunning, nil
}

func (h *Hetzner) Stop(ctx context.Context, name string) error {
	server, err := h.GetByName(ctx, name)
	if err != nil {
		return err
	}
	if server == nil {
		return nil
	}

	_, _, err = h.client.Server.DeleteWithResult(ctx, server)

	return err
}

func (h *Hetzner) volumeByName(ctx context.Context, name string) (*hcloud.Volume, error) {
	volumes, _, err := h.client.Volume.List(ctx, hcloud.VolumeListOpts{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	volLen := len(volumes)
	if volLen > 1 {
		return nil, ErrMultipleVolumesFound(name)
	}
	if volLen == 0 {
		return nil, nil
	}

	return volumes[0], nil
}

func generateUserData(machineId, publicKey, volumeId string) (userData string, err error) {
	t, err := template.New("cloud-config.yaml").ParseFS(cloudConfig, "cloud-config.yaml")
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, map[string]string{
		"PublicKey": publicKey,
		"VolumeID":  volumeId,
	}); err != nil {
		return
	}

	userData = buf.String()

	return
}
