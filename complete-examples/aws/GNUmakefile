VENV := .venv
VENV_STAMP := $(VENV)/stamp
ACTIVATE := . $(VENV)/bin/activate

bin: bin/ansible-builder bin/ansible-navigator

bin/ansible-builder: $(VENV_STAMP)

bin/ansible-navigator: $(VENV_STAMP)

$(VENV_STAMP): requirements.txt
	test -d $(VENV_STAMP) || python3 -qm venv $(VENV)
	$(ACTIVATE); pip install -qU pip setuptools wheel
	$(ACTIVATE); pip install -qr requirements.txt
	touch $(VENV_STAMP)

build: build/ee

build/ee: bin/ansible-builder
	$(ACTIVATE); ansible-builder build --container-runtime=docker -t ansible-execution-env-aws-example:v1 --no-cache

.PHONY: bin bin/ansile-builder bin/ansible-navigator build build/ee
