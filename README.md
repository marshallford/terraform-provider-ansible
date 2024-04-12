# Terraform Provider Ansible

[![Registry](https://img.shields.io/badge/ansible-Terraform%20Registry-blue)](https://registry.terraform.io/providers/marshallford/ansible/latest/docs)
[![Go Report Card](https://goreportcard.com/badge/github.com/marshallford/terraform-provider-ansible)](https://goreportcard.com/report/github.com/marshallford/terraform-provider-ansible)

Run Ansible playbooks within Ansible execution environments (EE) using Terraform.

## Features

1. Run Ansible playbooks against Terraform managed infrastructure (without the `local-exec` provisioner). Eliminates the need for additional scripting or pipeline steps.
2. Construct Ansible inventories using other data sources and resources. Set Ansible host and group variables to values and secrets from other providers.
3. Utilize Ansible [execution environments](https://ansible.readthedocs.io/en/latest/getting_started_ee/index.html) (containers images) to customize and run the Ansible software stack. Isolate Ansible and its related dependencies (Python/System packages, collections, etc) to simplify pipeline and workstation setup.
4. Write JSONPath queries against [playbook artifacts](https://access.redhat.com/documentation/en-us/red_hat_ansible_automation_platform/2.0-ea/html/ansible_navigator_creator_guide/assembly-troubleshooting-navigator_ansible-navigator#proc-review-artifact_troubleshooting-navigator). Extract values from the playbook run for use elsewhere in the configuration. Examples include: Ansible facts, remote file contents, task results -- the possibilities are endless!
5. Control playbook re-run behavior using several "lifecycle" options, including a flag for running the playbook on resource destruction. Implement conditional plays/tasks with the environment variable `ANSIBLE_TF_OPERATION`.

## Complete Examples

* [AWS](./complete-examples/aws/)
* [Libvirt](./complete-examples/libvirt/)

## Support Matrix

> [!WARNING]
> Windows builds of this provider are unlikely to work. Bug reports and PRs are welcome.

> [!WARNING]
> All versions released prior to `v1.0.0` are to be considered [breaking changes](https://semver.org/#how-do-i-know-when-to-release-100).

|  Release | Ansible Navigator | Terraform |
|:--------:|:-----------------:|:---------:|
| < v1.0.0 |      >= 3.5.0     |  >= 1.6.0 |

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
