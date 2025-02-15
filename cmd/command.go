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

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/loft-sh/devpod/pkg/ssh"
	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/hetzner"
	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/options"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// commandCmd represents the command command
var commandCmd = &cobra.Command{
	Use:   "command",
	Short: "Run a command on the instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		options, err := options.FromEnv(false)
		if err != nil {
			return err
		}

		ctx := context.Background()

		command := os.Getenv("COMMAND")
		if command == "" {
			return fmt.Errorf("command environment variable is missing")
		}

		// Get private key
		privateKey, err := ssh.GetPrivateKeyRawBase(options.MachineFolder)
		if err != nil {
			return fmt.Errorf("load private key: %w", err)
		}

		// Create SSH client
		server, err := hetzner.NewHetzner(options.Token).GetByName(ctx, options.MachineID)
		if err != nil {
			return err
		} else if server == nil {
			return fmt.Errorf("vm not found")
		}

		// Call external address
		sshClient, err := ssh.NewSSHClient(hetzner.SSHUsername, fmt.Sprintf("%s:%d", server.PublicNet.IPv4.IP, hetzner.SSHPort), privateKey)
		if err != nil {
			return errors.Wrap(err, "create ssh client")
		}
		defer func() {
			err = sshClient.Close()
		}()

		// Run command
		if err := ssh.Run(ctx, sshClient, command, os.Stdin, os.Stdout, os.Stderr, nil); err != nil {
			return err
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(commandCmd)
}
