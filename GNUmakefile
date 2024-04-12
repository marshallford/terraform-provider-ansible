default: test

DOCKER_FLAGS += --rm
ifeq ($(shell tty > /dev/null && echo 1 || echo 0), 1)
DOCKER_FLAGS += -i
endif

DOCKER := docker
DOCKER_RUN := $(DOCKER) run $(DOCKER_FLAGS)

EDITORCONFIG_CHECKER_VERSION ?= 2.7.2
EDITORCONFIG_CHECKER := $(DOCKER_RUN) -v=$(CURDIR):/check docker.io/mstruebing/editorconfig-checker:$(EDITORCONFIG_CHECKER_VERSION)

YAMLLINT_VERSION ?= 0.31.0
YAMLLINT := $(DOCKER_RUN) -v=$(CURDIR):/code docker.io/pipelinecomponents/yamllint:$(YAMLLINT_VERSION) yamllint

GOLANGCI_LINT_VERSION ?= 1.57.2
GOLANGCI_LINT := $(DOCKER_RUN) -v=$(CURDIR):/code -w /code docker.io/golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION) golangci-lint run

VENV := .venv
VENV_STAMP := $(VENV)/stamp
ACTIVATE := . $(VENV)/bin/activate

lint: lint/editorconfig lint/yamllint lint/go

lint/editorconfig:
	$(EDITORCONFIG_CHECKER)

lint/yamllint:
	$(YAMLLINT) .

lint/go:
	$(GOLANGCI_LINT)

install:
	go install

test: test/pkg test/acc

test/pkg:
	go test ./pkg/... -v $(TESTARGS) -timeout 60m

test/acc:
	TF_ACC=1 go test ./internal/provider/... -v $(TESTARGS) -timeout 60m

docs:
	go generate ./...

deps: bin/ansible-navigator

bin/ansible-navigator: $(VENV_STAMP)

$(VENV_STAMP): requirements.txt
	test -d $(VENV_STAMP) || python3 -qm venv $(VENV)
	$(ACTIVATE); pip install -qU pip setuptools wheel
	$(ACTIVATE); pip install -qr requirements.txt
	touch $(VENV_STAMP)

.PHONY: lint lint/editorconfig lint/yamllint lint/go install test test/pkg test/acc docs deps bin/ansible-navigator
