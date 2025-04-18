default: python

VENV := .venv
VENV_STAMP := $(VENV)/stamp
ACTIVATE := . $(VENV)/bin/activate

$(VENV_STAMP): requirements.txt
	test -d $(VENV_STAMP) || python3 -qm venv $(VENV)
	$(ACTIVATE); pip install -qU pip setuptools wheel
	$(ACTIVATE); pip install -qr requirements.txt
	touch $(VENV_STAMP)

.PHONY: python
python: $(VENV_STAMP)

.PHONY: build build/eei
build: build/eei

build/eei: python
	$(ACTIVATE); ansible-builder build --container-runtime=docker -t ansible-execution-env-aws-example:v1 --no-cache
