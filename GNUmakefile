default: test

DOCKER_FLAGS += --rm
ifeq ($(shell tty > /dev/null && echo 1 || echo 0), 1)
DOCKER_FLAGS += -i
endif

DOCKER := docker
DOCKER_RUN := $(DOCKER) run $(DOCKER_FLAGS)
DOCKER_PULL := $(DOCKER) pull -q

TERRAFORM_VERSION ?= 1.14.0

EDITORCONFIG_CHECKER_VERSION ?= 3.4.1
EDITORCONFIG_CHECKER_IMAGE ?= docker.io/mstruebing/editorconfig-checker:v$(EDITORCONFIG_CHECKER_VERSION)
EDITORCONFIG_CHECKER := $(DOCKER_RUN) -v=$(CURDIR):/check $(EDITORCONFIG_CHECKER_IMAGE)

SHELLCHECK_VERSION ?= 0.11.0
SHELLCHECK_IMAGE ?= docker.io/koalaman/shellcheck:v$(SHELLCHECK_VERSION)
SHELLCHECK := $(DOCKER_RUN) -v=$(CURDIR):/mnt $(SHELLCHECK_IMAGE)

YAMLLINT_VERSION ?= 0.35.6
YAMLLINT_IMAGE ?= docker.io/pipelinecomponents/yamllint:$(YAMLLINT_VERSION)
YAMLLINT := $(DOCKER_RUN) -v=$(CURDIR):/code $(YAMLLINT_IMAGE) yamllint

GOLANGCI_VERSION ?= 2.5.0
GOLANGCI_IMAGE ?= docker.io/golangci/golangci-lint:v$(GOLANGCI_VERSION)
GOLANGCI := $(DOCKER_RUN) -v=$(CURDIR):/code -w /code $(GOLANGCI_IMAGE) golangci-lint run

UV ?= uv
VENV := .venv
VENV_STAMP := $(VENV)/stamp
ACTIVATE := . $(VENV)/bin/activate

$(VENV_STAMP): pyproject.toml $(wildcard uv.lock)
	$(UV) venv $(VENV) --clear
	$(UV) sync
	touch $(VENV_STAMP)

.PHONY: python
python: $(VENV_STAMP)

.PHONY: pull pull/editorconfig pull/shellcheck pull/yamllint pull/golangci
pull: pull/editorconfig pull/shellcheck pull/yamllint pull/golangci

pull/editorconfig:
	$(DOCKER_PULL) $(EDITORCONFIG_CHECKER_IMAGE)

pull/shellcheck:
	$(DOCKER_PULL) $(SHELLCHECK_IMAGE)

pull/yamllint:
	$(DOCKER_PULL) $(YAMLLINT_IMAGE)

pull/golangci:
	$(DOCKER_PULL) $(GOLANGCI_IMAGE)

.PHONY: lint lint/terraform lint/editorconfig lint/shellcheck lint/yamllint lint/golangci lint/ansible
lint: lint/terraform lint/editorconfig lint/shellcheck lint/yamllint lint/golangci lint/ansible

lint/terraform:
	terraform fmt -recursive -check

lint/editorconfig:
	$(EDITORCONFIG_CHECKER)

lint/shellcheck:
	$(SHELLCHECK) $(shell find . -type f -not -path '*/\$(VENV)/*' -name '*.sh')

lint/yamllint:
	$(YAMLLINT) .

lint/golangci:
	$(GOLANGCI)

lint/ansible: $(VENV_STAMP)
	$(ACTIVATE); ansible-lint docs examples

.PHONY: install cover docs

install:
	go install

cover:
	go tool cover -html=cover.out

docs:
	TFENV_TERRAFORM_VERSION=$(TERRAFORM_VERSION) go generate ./...

.PHONY: test test/docs test/pkg test/acc
test: test/docs test/pkg test/acc

test/docs:
	TFENV_TERRAFORM_VERSION=$(TERRAFORM_VERSION) go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate

test/pkg: $(VENV_STAMP)
	GOTOOLCHAIN=go1.25.4+auto go test ./pkg/... -v -coverprofile=cover.out $(TESTARGS) -timeout 60m

test/acc: $(VENV_STAMP)
	GOTOOLCHAIN=go1.25.4+auto TF_ACC=1 TFENV_TERRAFORM_VERSION=$(TERRAFORM_VERSION) go test ./internal/provider/... -v -coverprofile=cover.out $(TESTARGS) -timeout 60m
