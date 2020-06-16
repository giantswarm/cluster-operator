# DO NOT EDIT. Generated with:
#
#    devctl gen makefile
#

APPLICATION    := $(shell basename $(shell go list .))
BUILDTIMESTAMP := $(shell date -u '+%FT%TZ')
GITSHA1        := $(shell git rev-parse --verify HEAD)
OS             := $(shell go env GOOS)
SOURCES        := $(shell find . -name '*.go')
VERSION        := $(shell architect project version)
LDFLAGS        ?= -w -linkmode 'auto' -extldflags '-static' \
  -X '$(shell go list .)/pkg/project.buildTimestamp=${BUILDTIMESTAMP}' \
  -X '$(shell go list .)/pkg/project.gitSHA=${GITSHA1}' \
  -X '$(shell go list .)/pkg/project.version=${VERSION}'
.DEFAULT_GOAL := build

.PHONY: build build-darwin build-linux
## build: builds a local binary
build: $(APPLICATION)
## build-darwin: builds a local binary for darwin/amd64
build-darwin: $(APPLICATION)-darwin
## build-linux: builds a local binary for linux/amd64
build-linux: $(APPLICATION)-linux

$(APPLICATION): $(APPLICATION)-$(OS)
	cp -a $< $@

$(APPLICATION)-%: $(SOURCES)
	GOOS=$* GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $@ .

.PHONY: install
## install: install the application
install:
	go install -ldflags "$(LDFLAGS)" .

.PHONY: run
## run: runs go run main.go
run:
	go run -ldflags "$(LDFLAGS)" -race .

.PHONY: clean
## clean: cleans the binary
clean:
	rm -f $(APPLICATION)*
	go clean

.PHONY: lint
## lint: runs golangci-lint
lint:
	golangci-lint run -E gosec -E goconst --timeout=15m ./...

.PHONY: test
## test: runs go test with default values
test:
	go test -ldflags "$(LDFLAGS)" -race ./...

.PHONY: build-docker
## build-docker: builds docker image to registry
build-docker: build-linux
	docker build -t ${APPLICATION}:${VERSION} .

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
