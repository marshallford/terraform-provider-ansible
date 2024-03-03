# Terraform Provider Ansible

[![Registry](https://img.shields.io/badge/ansible-Terraform%20Registry-blue)](https://registry.terraform.io/providers/marshallford/ansible/latest/docs)

Run Ansible playbooks within an Ansible execution environment (EE) using Terraform managed inventories/hosts.

> [!WARNING]
> All versions released prior to `v1.0.0` are to be considered [breaking changes](https://semver.org/#how-do-i-know-when-to-release-100).

> [!WARNING]
> Windows builds of this provider are unlikely to work. Bug reports and PRs are welcome.

## Support Matrix

|  Release | Ansible Navigator | Terraform |
|:--------:|:-----------------:|:---------:|
| < v1.0.0 |      >= 3.5.0     |  >= 1.5.0 |

## Development Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads)
- [Go](https://golang.org/doc/install)
- [Ansible Navigator](https://ansible.readthedocs.io/projects/navigator/installation/)
- [Docker](https://docs.docker.com/engine/install/) or [Podman](https://podman.io/docs/installation)

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#development-requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make docs`.

In order to run the full suite of Acceptance tests, run `make test/acc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make test/acc
```
