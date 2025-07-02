# DevPod Provider Hetzner

<!-- markdownlint-disable-next-line MD013 MD034 -->
[![Go Report Card](https://goreportcard.com/badge/github.com/mrsimonemms/devpod-provider-hetzner)](https://goreportcard.com/report/github.com/mrsimonemms/devpod-provider-hetzner)

DevPod on Hetzner

<!-- toc -->

* [Usage](#usage)
* [Development](#development)
  * [Required environment variables](#required-environment-variables)
  * [Testing independently of DevPod](#testing-independently-of-devpod)
  * [Testing in the DevPod ecosystem](#testing-in-the-devpod-ecosystem)
* [Contributing](#contributing)
  * [Open in a container](#open-in-a-container)
  * [Commit style](#commit-style)

<!-- Regenerate with "pre-commit run -a markdown-toc" -->

<!-- tocstop -->

> Use [this referral code](https://hetzner.cloud/?ref=UWVUhEZNkm6p) to get €20 in
> credits (at time of writing).

[DevPod](https://devpod.sh/) on [Hetzner](https://hetzner.cloud/?ref=UWVUhEZNkm6p).
This is based upon the [DigitalOcean provider](https://github.com/loft-sh/devpod-provider-digitalocean).

## Usage

To use this provider in your DevPod setup, you will need to do the following steps:

1. See the [DevPod documentation](https://devpod.sh/docs/managing-providers/add-provider)
   for how to add a provider
1. Use the reference `mrsimonemms/devpod-provider-hetzner` to download the latest
   release from GitHub
1. Get an [API token](https://docs.hetzner.com/cloud/api/getting-started/generating-api-token/)
   from Hetzner with `read & write` access. This will be used to manage resources.

## Development

### Required environment variables

> These are pre-configured in Dev Containers

| Variable | Description | Example |
| --- | --- | --- |
| `DISK_IMAGE` | Hetzner image tag | `docker-ce` |
| `DISK_SIZE` | Disk size in GB | `30` |
| `GIT_REPO` | Git repo to download | `github.com/mrsimonemms/devpod-provider-hetzner` |
| `HCLOUD_TOKEN` | [Hetzner API token](https://docs.hetzner.com/cloud/api/getting-started/generating-api-token/) with `read & write` access | - |
| `MACHINE_FOLDER` | Local home folder | `~/.ssh` |
| `MACHINE_ID` | Unique identifier for the machine | `some-machine-id` |
| `MACHINE_TYPE` | Hetzner machine size | `cx22` |
| `REGION` | Hetzner region ID | `nbg1` |
| `TOKEN` | **Deprecated**. Replaced by `HCLOUD_TOKEN` | - |

### Testing independently of DevPod

To test the provider workflow, you can run the CLI commands directly.

| Command | Description | Example |
| --- | --- | --- |
| `command` | Run a command on the instance | `COMMAND="ls -la" go run . command` |
| `create` | Create an instance | `go run . create` |
| `delete` | Delete an instance and volume | `go run . delete` |
| `init` | Initialise an instance | `go run . init` |
| `start` | Start an instance | `go run . start` |
| `status` | Retrieve the status of an instance | `go run . status` |
| `stop` | Stop an instance | `go run . stop` |

### Testing in the DevPod ecosystem

> This assumes a Linux AMD64 workspace - if you're developing on any other machine
> please update the instructions for that machine (PRs welcome).
>
> These paths may differ on your machine.

To test the provider within the DevPod ecosystem:

1. Install the latest version of the [Hetzner provider](#usage)
1. Backup the original binary:

   ```shell
   mv ~/.devpod/contexts/default/providers/hetzner/binaries/hetzner_provider/devpod-provider-hetzner-linux-amd64 ~/.devpod/contexts/default/providers/hetzner/binaries/hetzner_provider/devpod-provider-hetzner-linux-amd64-orig
   ```

1. Build the binary:

   ```shell
   go build .
   ```

1. Move the new binary to the DevPod base:

   ```shell
   mv ./devpod-provider-hetzner ~/.devpod/contexts/default/providers/hetzner/binaries/hetzner_provider/devpod-provider-hetzner-linux-amd64
   ```

## Contributing

* Get a [Hetzner](https://hetzner.cloud/?ref=UWVUhEZNkm6p) account
* Get an [API token](https://docs.hetzner.com/cloud/api/getting-started/generating-api-token/)
  with `read & write` access
* Save this as `HCLOUD_TOKEN` in your `.envrc` file

### Open in a container

* [Open in a container](https://code.visualstudio.com/docs/devcontainers/containers)

### Commit style

All commits must be done in the [Conventional Commit](https://www.conventionalcommits.org)
format.

```git
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```
