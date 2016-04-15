PKGS := $(shell go list ./... | grep -v example)
GOLINT := $(GOPATH)/bin/golint
.PHONY: all test

GOVERSION := $(shell go version | grep 1.6)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.6)
endif
export GO15VENDOREXPERIMENT = 1

all: test

test: $(PKGS)

$(PKGS): $(GOLINT)
	@echo "FORMATTING..."
	@gofmt -w=true $(GOPATH)/src/$@/*.go
	@echo "LINTING..."
	@$(GOLINT) $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "VETTING..."
	@go vet $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "TESTING..."
	@TEST_MONGO_URL=$(TEST_MONGO_URL) go test -v $@
	@echo ""

$(GOLINT):
	go get github.com/golang/lint/golint
