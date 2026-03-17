AGENT_CLAUDE_IMAGE ?= claude-code
AGENT_CLAUDE_STATE_MOUNT ?= claude_state:/home/agent/.claude
AGENT_CLAUDE_CONFIG_SOURCE ?= $(HOME)/.config/agent/claude/.claude.json
AGENT_CLAUDE_CONFIG_TARGET = /home/agent/.claude.json

AGENT_CODEX_IMAGE ?= codex-cli
AGENT_CODEX_STATE_MOUNT ?= codex_state:/home/node/.codex

AGENT_NO_PROXY ?= host.docker.internal,localhost
AGENT_DOCKER_ADD_HOST ?= host.docker.internal:host-gateway

.PHONY: init init-claude init-codex build-claude run-claude shell-claude build-codex run-codex shell-codex

define require_var
	@if [ -z "$($1)" ]; then echo "error: required env var $1 is not set"; exit 2; fi
endef

define require_file
	@if [ ! -f "$($1)" ]; then echo "error: required file $($1) does not exist"; exit 2; fi
endef

define ensure_volume
	@docker volume inspect $1 >/dev/null 2>&1 || docker volume create $1 >/dev/null
endef

state_volume = $(firstword $(subst :, ,$1))

init: init-claude init-codex

init-claude:
	$(call ensure_volume,$(call state_volume,$(AGENT_CLAUDE_STATE_MOUNT)))

init-codex:
	$(call ensure_volume,$(call state_volume,$(AGENT_CODEX_STATE_MOUNT)))

build-claude:
	$(call require_var,AGENT_ALL_PROXY)
	docker build \
		--add-host=$(AGENT_DOCKER_ADD_HOST) \
		--build-arg ALL_PROXY=$(AGENT_ALL_PROXY) \
		--build-arg NO_PROXY=$(AGENT_NO_PROXY) \
		-t $(AGENT_CLAUDE_IMAGE) \
		./claude

run-claude:
	$(call require_var,AGENT_ALL_PROXY)
	$(call require_var,AGENT_CLAUDE_CONFIG_SOURCE)
	$(call require_file,AGENT_CLAUDE_CONFIG_SOURCE)
	docker run -it --rm \
		--add-host=$(AGENT_DOCKER_ADD_HOST) \
		-v $(AGENT_CLAUDE_STATE_MOUNT) \
		-v $(AGENT_CLAUDE_CONFIG_SOURCE):$(AGENT_CLAUDE_CONFIG_TARGET):rw \
		-v ../:/app \
		-e ALL_PROXY=$(AGENT_ALL_PROXY) \
		-e NO_PROXY=$(AGENT_NO_PROXY) \
		$(AGENT_CLAUDE_IMAGE)

shell-claude:
	$(call require_var,AGENT_ALL_PROXY)
	$(call require_var,AGENT_CLAUDE_CONFIG_SOURCE)
	$(call require_file,AGENT_CLAUDE_CONFIG_SOURCE)
	docker run -it --rm \
		--add-host=$(AGENT_DOCKER_ADD_HOST) \
		-v $(AGENT_CLAUDE_STATE_MOUNT) \
		-v $(AGENT_CLAUDE_CONFIG_SOURCE):$(AGENT_CLAUDE_CONFIG_TARGET):rw \
		-v ../:/app \
		-e ALL_PROXY=$(AGENT_ALL_PROXY) \
		-e NO_PROXY=$(AGENT_NO_PROXY) \
		$(AGENT_CLAUDE_IMAGE) bash

build-codex:
	$(call require_var,AGENT_ALL_PROXY)
	docker build --no-cache -t $(AGENT_CODEX_IMAGE) ./codex

run-codex:
	$(call require_var,AGENT_ALL_PROXY)
	docker run -it --rm \
		--add-host=$(AGENT_DOCKER_ADD_HOST) \
		-v $(AGENT_CODEX_STATE_MOUNT) \
		-v ../:/app \
		-e ALL_PROXY="$(AGENT_ALL_PROXY)" \
		-e NO_PROXY="$(AGENT_NO_PROXY)" \
		$(AGENT_CODEX_IMAGE)

shell-codex:
	$(call require_var,AGENT_ALL_PROXY)
	docker run -it --rm \
		--add-host=$(AGENT_DOCKER_ADD_HOST) \
		-v $(AGENT_CODEX_STATE_MOUNT) \
		-v ../:/app \
		-e ALL_PROXY="$(AGENT_ALL_PROXY)" \
		-e NO_PROXY="$(AGENT_NO_PROXY)" \
		$(AGENT_CODEX_IMAGE)
