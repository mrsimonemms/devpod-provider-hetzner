#!/usr/bin/env bash
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

export GO111MODULE=on
# export GOFLAGS=-mod=vendor

PROVIDER_ROOT=$(git rev-parse --show-toplevel)
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
DATE=$(date "+%Y-%m-%d")
BUILD_PLATFORM=$(uname -a | awk '{print tolower($1);}')

echo "Current working directory is $(pwd)"
echo "PATH is $PATH"
echo "GOPATH is $GOPATH"

if [[ "$(pwd)" != "${PROVIDER_ROOT}" ]]; then
  echo "you are not in the root of the repo" 1>&2
  echo "please cd to ${PROVIDER_ROOT} before running this script" 1>&2
  exit 1
fi

GO_BUILD_CMD="go build"
GO_BUILD_LDFLAGS="-w -s -X ${GIT_REPO}/cmd.Version=${RELEASE_VERSION} -X ${GIT_REPO}/cmd.GitCommit=${GIT_COMMIT}"

if [[ -z "${PROVIDER_BUILD_PLATFORMS}" ]]; then
    PROVIDER_BUILD_PLATFORMS="linux windows darwin"
fi

if [[ -z "${PROVIDER_BUILD_ARCHS}" ]]; then
    PROVIDER_BUILD_ARCHS="amd64 arm64"
fi

# Create the dist directory
mkdir -p "${PROVIDER_ROOT}/dist"

for OS in ${PROVIDER_BUILD_PLATFORMS[@]}; do
  for ARCH in ${PROVIDER_BUILD_ARCHS[@]}; do
    NAME="devpod-provider-hetzner-${OS}-${ARCH}"
    if [[ "${OS}" == "windows" ]]; then
      NAME="${NAME}.exe"
    fi

    # darwin 386 is deprecated and shouldn't be used anymore
    if [[ "${ARCH}" == "386" && "${OS}" == "darwin" ]]; then
        echo "Building for ${OS}/${ARCH} not supported."
        continue
    fi

    # arm64 build is only supported for darwin
    if [[ "${ARCH}" == "arm64" && "${OS}" == "windows" ]]; then
        echo "Building for ${OS}/${ARCH} not supported."
        continue
    fi

    echo "Building for ${OS}/${ARCH}"
    GOARCH=${ARCH} GOOS=${OS} CGO_ENABLED=0 ${GO_BUILD_CMD} -ldflags "${GO_BUILD_LDFLAGS}"\
      -o "${PROVIDER_ROOT}/dist/${NAME}" main.go
    shasum -a 256 "${PROVIDER_ROOT}/dist/${NAME}" | cut -d ' ' -f 1 > "${PROVIDER_ROOT}/dist/${NAME}".sha256
  done
done

# generate provider.yaml
go run "${PROVIDER_ROOT}/hack/provider/main.go" ${RELEASE_VERSION} > "${PROVIDER_ROOT}/dist/provider.yaml"
