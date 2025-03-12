SHELL := bash

################################################
# STEP 1: Fail if .env is missing
################################################
ifeq (,$(wildcard .env))
  $(error ".env file is missing! Please create a .env with HUB_USERNAME, REPO_NAME, VERSION, etc.")
endif

################################################
# STEP 2: Include .env, then export the vars
################################################
include .env
export $(shell sed 's/=.*//' .env)

################################################
# STEP 3: Fallback defaults (if not in .env)
################################################
HUB_USERNAME ?= 
REPO_NAME ?= 
VERSION ?=

# For multi-arch builds
PLATFORMS=linux/amd64,linux/arm64

################################################
# STEP 4: A special 'check-env' target that
#         fails if any required variable is empty
################################################
.PHONY: check-env
check-env:
ifeq ($(strip $(HUB_USERNAME)),)
	$(error "HUB_USERNAME is not set! Please set it in .env")
endif
ifeq ($(strip $(REPO_NAME)),)
	$(error "REPO_NAME is not set! Please set it in .env")
endif
ifeq ($(strip $(VERSION)),)
	$(error "VERSION is not set! Please set it in .env")
endif

################################################
# STEP 5: Normal build targets, each depends
#         on 'check-env' to ensure vars are set
################################################

build: check-env
	docker build -f Dockerfile \
		-t $(HUB_USERNAME)/$(REPO_NAME):$(VERSION) .

push: check-env
	docker push $(HUB_USERNAME)/$(REPO_NAME):$(VERSION)

run: check-env
	docker run -d -p 8080:8080 --cpus=2 --name go-load-lab \
		$(HUB_USERNAME)/$(REPO_NAME):$(VERSION)

buildx: check-env
	docker buildx build \
		--platform $(PLATFORMS) \
		-f Dockerfile \
		-t $(HUB_USERNAME)/$(REPO_NAME):$(VERSION) \
		--push .
