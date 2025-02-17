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
	"os"

	"github.com/loft-sh/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "devpod-provider-hetzner",
	Short:   "DevPod on Hetzner",
	Version: Version,
	PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
		log.Default.MakeRaw()

		logLevel := os.Getenv("DEVPOD_LOG_LEVEL")
		if logLevel != "" {
			if lvl, err := logrus.ParseLevel(logLevel); err != nil {
				log.Default.Error(errors.Wrap(err, "invalid log level provided, continuing"))
			} else {
				log.Default.SetLevel(lvl)
			}
		}

		if token := os.Getenv("TOKEN"); token != "" {
			log.Default.Warn("TOKEN envvar is deprecated in favour of HCLOUD_TOKEN")
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
