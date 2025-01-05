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
	"time"

	"github.com/loft-sh/devpod/pkg/client"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/hetzner"
	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/options"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop an instance",
	RunE: func(_ *cobra.Command, args []string) error {
		options, err := options.FromEnv(false)
		if err != nil {
			return err
		}

		ctx := context.Background()

		hetznerClient := hetzner.NewHetzner(options.Token)

		err = hetznerClient.Stop(ctx, options.MachineID)
		if err != nil {
			return err
		}

		// Wait until it's stopped
		for {
			status, err := hetznerClient.Status(ctx, options.MachineID)
			if err != nil {
				log.Default.Errorf("Error retrieving server status: %v", err)
				break
			} else if status == client.StatusStopped {
				break
			}

			time.Sleep(time.Second)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
