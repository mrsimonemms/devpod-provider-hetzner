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
	"errors"
	"fmt"
)

var (
	ErrBadSSHKey            = errors.New("bad ssh key")
	ErrMultipleServersFound = func(name string) error {
		return fmt.Errorf("multiple server with name %s found", name)
	}
	ErrMultipleVolumesFound = func(name string) error {
		return fmt.Errorf("multiple volumes with name %s found", name)
	}
	ErrUnknownDiskImage = errors.New("unknown disk image")
	ErrUnknownMachineID = errors.New("unknown machine id")
	ErrUnknownRegion    = errors.New("unknown region")
)
