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
