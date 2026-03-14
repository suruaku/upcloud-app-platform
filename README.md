# upcloud-box

`upcloud-box` is a Go CLI for provisioning one secure Docker host on UpCloud and deploying one container workload.

## Features

- Initialize config, state, and cloud-init templates
- Provision an UpCloud server with cloud-init
- Verify SSH and Docker readiness after provisioning
- Deploy a single container with health checks and rollback
- Show infra + container + health status
- Destroy tracked server and clean local infra state
- One-command runtime flow with `up` (provision if needed, then deploy)

## Requirements

- Go 1.25+ (for local development)
- UpCloud account and API token
- Local SSH keypair (public key for cloud-init, private key for SSH access)

## Quickstart

1) Export your UpCloud token:

```bash
export UPCLOUD_TOKEN="ucat_..."
```

2) Initialize project files:

```bash
upcloud-box init --ssh-key ~/.ssh/id_ed25519.pub --user ubuntu
```

This creates:

- `upcloud-box.yaml`
- `.upcloud-box.state.json`
- `cloud-init.yaml`

3) Edit config values in `upcloud-box.yaml`.

4) Run full runtime flow:

```bash
upcloud-box up
```

5) Inspect status:

```bash
upcloud-box status
```

6) Clean up:

```bash
upcloud-box destroy --yes
```

## Core Commands

- `upcloud-box init` - scaffold config/state/cloud-init
- `upcloud-box provision` - create server and persist infra state
- `upcloud-box deploy` - deploy container and update deploy state
- `upcloud-box up` - provision if needed, then deploy
- `upcloud-box status` - local state + remote infra/container/health summary
- `upcloud-box destroy` - delete tracked server and clear infra state fields

## Useful Flags

- `--config <path>`: custom config path (default: `upcloud-box.yaml`)
- `--verbose`: show detailed error output and verbose logs
- `--no-spinner`: disable spinner progress output

## Troubleshooting

- `initialize provider failed (auth)`: verify `UPCLOUD_TOKEN` is set and valid.
- `... failed (quota)`: check UpCloud resource limits and selected zone capacity.
- `post-provision checks failed (ssh)`: confirm `ssh.user`, `ssh.private_key_path`, and cloud-init user/key setup match.
- `deploy container failed (health)`: verify app startup, exposed port mapping, and `deploy.healthcheck_url`.
- `status` shows server missing: run `upcloud-box up` to reprovision or `upcloud-box destroy --yes` to clean state.

## Release Artifacts

On version tags (`v*`), GitHub Actions builds and publishes binaries for:

- Linux: amd64, arm64
- macOS: amd64, arm64
- Windows: amd64
