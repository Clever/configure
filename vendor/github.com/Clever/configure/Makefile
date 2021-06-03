include golang.mk
.DEFAULT_GOAL := test

SHELL := /bin/bash
PKG := github.com/Clever/configure
PKGS := $(shell go list ./... | grep -v example)
.PHONY: all test

$(eval $(call golang-version-check,1.8))

all: test

test: $(PKGS)

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)
