#!/bin/bash
# Copyright 2023 Simon Emms <simon@simonemms.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

echo "Listing volumes"
volumes=$(hcloud volume list -l machineId=${MACHINE_ID} -o noheader -o columns=id)

for i in ${volumes}; do
  echo "Detaching volume ${i} from server"
  hcloud volume detach "${i}" || true

  echo "Deleting volume ${i} from server"
  hcloud volume delete "${i}"
done

echo "Listing servers"
servers=$(hcloud server list -l machineId=${MACHINE_ID} -o noheader -o columns=id)

for i in ${servers}; do
  echo "Deleting server ${i}"
  hcloud server delete "${i}"
done

echo "Listing SSH keys"
ssh_keys=$(hcloud ssh-key list -l machineId=${MACHINE_ID} -o noheader -o columns=id)

for i in ${ssh_keys}; do
  echo "Deleting SSH key ${i}"
  hcloud ssh-key delete "${i}"
done
