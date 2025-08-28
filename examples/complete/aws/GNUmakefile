default: python

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

.PHONY: build build/eei
build: build/eei

build/eei: $(VENV_STAMP)
	$(ACTIVATE); ansible-builder build --container-runtime=docker -t terraform-provider-ansible-example-aws:v1 --no-cache
