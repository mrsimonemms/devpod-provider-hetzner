/*
 * Copyright 2023 Simon Emms <simon@simonemms.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hetzner

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	cryptoSsh "golang.org/x/crypto/ssh"

	"github.com/google/uuid"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/loft-sh/devpod/pkg/client"
	"github.com/loft-sh/devpod/pkg/ssh"
	"github.com/loft-sh/log"
	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/options"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

//go:embed cloud-config.yaml
var cloudConfig embed.FS

type cloudInit struct {
	Status string `json:"status"`
}

type Hetzner struct {
	client *hcloud.Client
}

func NewHetzner(token string) *Hetzner {
	return &Hetzner{
		client: hcloud.NewClient(hcloud.WithToken(token)),
	}
}

func (h *Hetzner) BuildServerOptions(ctx context.Context, opts *options.Options) (*hcloud.ServerCreateOpts, *string, []byte, error) {
	log.Default.Debugf("Machine folder path: %s", opts.MachineFolder)

	publicKeyBase, err := ssh.GetPublicKeyBase(opts.MachineFolder)
	if err != nil {
		return nil, nil, nil, err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return nil, nil, nil, err
	}

	privateKey, err := ssh.GetPrivateKeyRawBase(opts.MachineFolder)
	if err != nil {
		return nil, nil, nil, err
	}

	location, _, err := h.client.Location.GetByName(ctx, opts.Region)
	if err != nil {
		return nil, nil, nil, err
	}
	if location == nil {
		return nil, nil, nil, ErrUnknownRegion
	}

	serverType, _, err := h.client.ServerType.GetByName(ctx, opts.MachineType)
	if err != nil {
		return nil, nil, nil, err
	}
	if serverType == nil {
		return nil, nil, nil, ErrUnknownMachineID
	}

	fingerprint, err := generateSSHKeyFingerprint(string(publicKey))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to generate fingerprint for public ssh key")
	}

	sshKey, _, err := h.client.SSHKey.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		return nil, nil, nil, err
	}

	if sshKey == nil {
		// Generate name
		machineId := opts.MachineID
		if len(machineId) >= 24 {
			machineId = opts.MachineID[:24]
		}
		name := fmt.Sprintf("%s-%s", machineId, uuid.NewString()[:8])

		log.Default.Infof("Uploading SSH key: %s", name)

		// Upload the key
		uploadedSSHKey, _, err := h.client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
			Name:      name,
			PublicKey: string(publicKey),
			Labels: map[string]string{
				"type":         "devpod",
				labelMachineId: opts.MachineID,
			},
		})
		if err != nil {
			return nil, nil, nil, err
		}

		sshKey = uploadedSSHKey
	}

	arch := hcloud.ArchitectureX86
	if strings.HasPrefix(opts.MachineType, "cax") {
		// Machines starting "cax" are ARM64
		arch = hcloud.ArchitectureARM
	}
	image, _, err := h.client.Image.GetByNameAndArchitecture(ctx, opts.DiskImage, arch)
	if err != nil {
		return nil, nil, nil, err
	}
	if image == nil {
		return nil, nil, nil, ErrUnknownDiskImage
	}

	return &hcloud.ServerCreateOpts{
		Name:       opts.MachineID,
		Location:   location,
		ServerType: serverType,
		Image:      image,
		Labels: map[string]string{
			"type": "devpod",
		},
		SSHKeys: []*hcloud.SSHKey{
			sshKey,
		},
	}, hcloud.Ptr(string(publicKey)), privateKey, nil
}

func (h *Hetzner) Create(ctx context.Context, req *hcloud.ServerCreateOpts, diskSize int, publicKey string, privateKeyFile []byte) error {
	log.Default.Info("Creating DevPod instance")

	volume, err := h.volumeByName(ctx, req.Name)
	if err != nil {
		return err
	}

	if volume == nil {
		// Create the volume as it doesn't exist
		log.Default.Info("Creating a new volume")

		result, _, err := h.client.Volume.Create(ctx, hcloud.VolumeCreateOpts{
			Location:  req.Location,
			Name:      req.Name,
			Size:      diskSize,
			Format:    hcloud.Ptr("ext4"),
			Automount: hcloud.Ptr(false),
			Labels: map[string]string{
				"type": "devpod",
			},
		})
		if err != nil {
			return err
		}

		actions := append([]*hcloud.Action{result.Action}, result.NextActions...)

		for _, a := range actions {
			log.Default.Debugf("Waiting for action to complete: %s", a.Command)
			if err := h.waitForActionCompletion(ctx, a); err != nil {
				log.Default.Errorf("Error in volume creation action: %s, %s", a.Command, err)
				return err
			}
		}

		log.Default.Info("Volume successfully created")

		volume = result.Volume
	}

	// Generate the config init
	userData, err := generateUserData(req.Name, publicKey, volume.ID)
	if err != nil {
		return err
	}
	// Add to server config
	req.UserData = userData.String()

	// Add volume to the server config
	req.Volumes = []*hcloud.Volume{
		{
			ID: volume.ID,
		},
	}

	// Create the server
	log.Default.Info("Creating a new server")
	server, _, err := h.client.Server.Create(ctx, *req)
	if err != nil {
		return err
	}

	log.Default.Info("Server creation triggered")

	actions := append([]*hcloud.Action{server.Action}, server.NextActions...)

	for _, a := range actions {
		log.Default.Debugf("Waiting for action to complete: %s", a.Command)
		if err := h.waitForActionCompletion(ctx, a, time.Minute*5); err != nil {
			log.Default.Errorf("Error in server creation action: %s, %s", a.Command, err)
			return err
		}
	}

	log.Default.Info("Server created - provisioning")

	attempt := 0

	for {
		if attempt >= maxServerConnectAttempts {
			return fmt.Errorf("exceeded attempts to connect to server: %d", attempt)
		}
		attempt++
		log.Default.Debugf("Attempt %d of %d", attempt, maxServerConnectAttempts)

		time.Sleep(time.Second)

		log.Default.Debug("Checking server provision status")

		// Check the server is provisioned - this runs "ssh user@path cloud-init status"
		sshClient, err := ssh.NewSSHClient(SSH_USERNAME, fmt.Sprintf("%s:%d", server.Server.PublicNet.IPv4.IP, SSH_PORT), privateKeyFile)
		if err != nil {
			log.Default.Warnf("Unable to connect to server: %v", err)
			continue
		}
		defer func() {
			err = sshClient.Close()
		}()

		buf := new(bytes.Buffer)
		if err := ssh.Run(ctx, sshClient, "cloud-init status || true", &bytes.Buffer{}, buf, &bytes.Buffer{}, nil); err != nil {
			log.Default.Errorf("Error retrieving cloud-init status, %v", err)
			continue
		}

		var status cloudInit
		if err := yaml.Unmarshal(buf.Bytes(), &status); err != nil {
			log.Default.Errorf("Unable to parse cloud-init YAML: %v", err)
			continue
		}

		if status.Status == "done" {
			// The server is ready
			break
		}

		log.Default.Debug("Server not yet provisioned")
	}

	log.Default.Info("Server provisioned")

	return nil
}

func (h *Hetzner) Delete(ctx context.Context, name string) error {
	// Delete SSH key
	keys, _, err := h.client.SSHKey.List(ctx, hcloud.SSHKeyListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: fmt.Sprintf("%s=%s", labelMachineId, name),
		},
	})
	if err != nil {
		return err
	}

	for _, k := range keys {
		log.Default.Infof("Deleting SSH key: %s", k.Name)
		_, err = h.client.SSHKey.Delete(ctx, k)
		if err != nil {
			return err
		}
	}

	// Delete volume
	volume, err := h.volumeByName(ctx, name)
	if err != nil {
		return err
	} else if volume != nil && volume.Server != nil {
		// Detatch volume
		action, _, err := h.client.Volume.Detach(ctx, volume)
		if err != nil {
			return errors.Wrap(err, "detach volume")
		}

		if err := h.waitForActionCompletion(ctx, action); err != nil {
			log.Default.Errorf("Error in volume detach action: %s, %s", action.Command, err)
			return err
		}
	}

	// Wait until the volume is detached
	for {
		time.Sleep(time.Second)

		// re-get volume
		volume, err = h.volumeByName(ctx, name)
		if err != nil {
			return err
		} else if volume == nil || volume.Server == nil {
			break
		}
	}

	// delete volume
	if volume != nil {
		_, err = h.client.Volume.Delete(ctx, volume)
		if err != nil {
			return errors.Wrap(err, "delete volume")
		}
	}

	server, err := h.GetByName(ctx, name)
	if err != nil {
		return err
	} else if server == nil {
		return nil
	}

	result, _, err := h.client.Server.DeleteWithResult(ctx, server)
	if err != nil {
		return err
	}

	return h.waitForActionCompletion(ctx, result.Action)
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

	result, _, err := h.client.Server.DeleteWithResult(ctx, server)
	if err != nil {
		return err
	}

	return h.waitForActionCompletion(ctx, result.Action)
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

// waitForActionCompletion if a command returns an *hcloud.Action struct, this is a long-running job on the Hetzner side.
// If we need to wait for it to finish to rely upon it later, this command will wait until we have success or that
// there's an error
func (h *Hetzner) waitForActionCompletion(ctx context.Context, action *hcloud.Action, timeout ...time.Duration) error {
	if action == nil {
		log.Default.Debug("No action received")
		return nil
	}

	if len(timeout) == 0 {
		timeout = []time.Duration{
			time.Minute,
		}
	}

	startTime := time.Now()
	timeoutTime := startTime.Add(timeout[0])

	for {
		time.Sleep(time.Second)

		now := time.Now()

		if now.After(timeoutTime) {
			log.Default.Errorf("action timedout: %d", action.ID)

			return fmt.Errorf("action timed out")
		}

		log.Default.Debug("Checking action status")

		status, _, err := h.client.Action.GetByID(ctx, action.ID)
		if err != nil {
			return err
		}

		log.Default.Debugf("Current status: %s", status.Status)

		if status.Status == hcloud.ActionStatusError {
			log.Default.Errorf("error completing action - code: %s, message: %s", status.ErrorCode, status.ErrorMessage)
			return fmt.Errorf("%s: %s", status.ErrorCode, status.ErrorMessage)
		}

		if status.Status == hcloud.ActionStatusSuccess {
			break
		}
	}

	return nil
}

func generateSSHKeyFingerprint(publicKey string) (string, error) {
	pk, _, _, _, err := cryptoSsh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return "", err
	}

	return cryptoSsh.FingerprintLegacyMD5(pk), nil
}

func generateUserData(_, publicKey string, volumeId int64) (*bytes.Buffer, error) {
	t, err := template.New("cloud-config.yaml").ParseFS(cloudConfig, "cloud-config.yaml")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, map[string]string{
		"PublicKey": strings.TrimSuffix(publicKey, "\n"),
		"VolumeID":  strconv.FormatInt(volumeId, 10),
		"Username":  SSH_USERNAME,
	}); err != nil {
		return nil, err
	}

	return buf, nil
}
