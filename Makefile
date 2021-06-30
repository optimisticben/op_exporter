SHELL := /bin/bash

VERSION := ${VERSION}
GITCOMMIT := `git rev-parse HEAD`
BUILDDATE := `date +%Y-%m-%d`
BUILDUSER := `whoami`

LDFLAGSSTRING :=-X version.Version=$(VERSION)
LDFLAGSSTRING +=-X version.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X version.BuildDate=$(BUILDDATE)
LDFLAGSSTRING +=-X version.BuildUser=$(BUILDUSER)

LDFLAGS :=-ldflags "$(LDFLAGSSTRING)"

.PHONY: all build

all: build

# Build binary
build:
	CGO_ENABLED=0 go build $(LDFLAGS)