ifndef VERBOSE
.SILENT:
endif

override CURRENT_DIR = $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
override DOCKER_MOUNT_SUFFIX ?= consistent

ifeq ($(GO111MODULE),auto)
override GO111MODULE = on
endif

ifeq ($(OS),Windows_NT)
override ROOT_DIR = $(shell echo $(CURRENT_DIR) | sed -e "s:^/./:\U&:g")
else
override ROOT_DIR = $(CURRENT_DIR)
endif

go-mockery: ## generate golang mock objects
	go get github.com/vektra/mockery/.../ ;\
	. ${ROOT_DIR}/scripts/mockery.sh ${ROOT_DIR}/scripts
.PHONY: go-mockery

init:
	. ${ROOT_DIR}/scripts/common.sh ${ROOT_DIR}/scripts ;\
	mkdir -p $${PROTO_GEN_PATH}
.PHONY: init

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

.DEFAULT_GOAL := help