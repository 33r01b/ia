AGENT_GEMINI_IMAGE ?= gemini-cli
AGENT_GEMINI_STATE_MOUNT ?= gemini_state:/home/node/.gemini
AGENT_GEMINI_ENV_MOUNT ?= /dev/null:/app/.env

AGENT_NO_PROXY ?= host.docker.internal,localhost
AGENT_DOCKER_ADD_HOST ?= host.docker.internal:host-gateway

.PHONY: init build run shell

define require_var
	@if [ -z "$($1)" ]; then echo "error: required env var $1 is not set"; exit 2; fi
endef

define ensure_volume
	@docker volume inspect $1 >/dev/null 2>&1 || docker volume create $1 >/dev/null
endef

state_volume = $(firstword $(subst :, ,$1))

init:
	$(call ensure_volume,$(call state_volume,$(AGENT_GEMINI_STATE_MOUNT)))

build:
	$(call require_var,AGENT_ALL_PROXY)
	docker build --no-cache -t $(AGENT_GEMINI_IMAGE) .

run:
	$(call require_var,AGENT_ALL_PROXY)
	docker run -it --rm \
	  --add-host=$(AGENT_DOCKER_ADD_HOST) \
	  -v $(AGENT_GEMINI_STATE_MOUNT) \
	  -v ./:/app \
	  -v $(AGENT_GEMINI_ENV_MOUNT) \
	  -e ALL_PROXY="$(AGENT_ALL_PROXY)" \
	  -e NO_PROXY="$(AGENT_NO_PROXY)" \
	  $(AGENT_GEMINI_IMAGE)

shell:
	$(call require_var,AGENT_ALL_PROXY)
	docker run -it --rm \
	  --add-host=$(AGENT_DOCKER_ADD_HOST) \
	  -v $(AGENT_GEMINI_STATE_MOUNT) \
	  -v ./:/app \
	  -v $(AGENT_GEMINI_ENV_MOUNT) \
	  -e ALL_PROXY="$(AGENT_ALL_PROXY)" \
	  -e NO_PROXY="$(AGENT_NO_PROXY)" \
	  $(AGENT_GEMINI_IMAGE)
