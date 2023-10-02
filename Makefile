include golang.mk
.DEFAULT_GOAL := test

SHELL := /bin/bash
PKG := github.com/Clever/configure
PKGS := $(shell go list ./... | grep -v example | grep -v /vendor)
.PHONY: all test

$(eval $(call golang-version-check,1.21))

all: test

test: $(PKGS)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)


install_deps:
	go mod vendor
