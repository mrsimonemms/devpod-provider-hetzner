/*
Copyright © 2023 Simon Emms <simon@simonemms.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/hetzner"
	"github.com/mrsimonemms/devpod-provider-hetzner/pkg/options"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Retrieve the status of an instance",
	RunE: func(_ *cobra.Command, args []string) error {
		options, err := options.FromEnv(false)
		if err != nil {
			return err
		}

		status, err := hetzner.NewHetzner(options.Token).Status(context.Background(), options.MachineID)
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(os.Stdout, status)
		return err
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
