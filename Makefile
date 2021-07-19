# Check for required command tools to build or stop immediately
EXECUTABLES = go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

# This should map to current directory
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# if you change the name of the binary, you need to edit docker/k8s/helm files too
BINARY=enrollment
# reads version number from VERSION file; VERSION file is also read by CI/CD process
VERSION=`cat ./VERSION` 
# Platforms to build for
PLATFORMS=linux
# CPU architectures to build for
ARCHITECTURES=amd64
# Enable CGO
CGOENABLED=1

# Setup linker flags option to embed version into binary (will be printed when binary starts)
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Debug/Info output...
$(info ROOT_DIR: $(ROOT_DIR))
$(info BINARY  : $(BINARY))
$(info VERSION : $(shell cat VERSION))

#---------------------------------------#
default: build

# builds native binary
build: 
	go build ${LDFLAGS} -o ${BINARY}

# cleans directory and then builds binary for all defined platforms and architectures
all: clean build_all

# builds binary for all defined platforms and architectures
build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); export CGO_ENABLED=$(CGOENABLED); go build $(LDFLAGS) -v -o $(BINARY)-$(GOOS)-$(GOARCH))))

# removes only binary files (no folders) we've created from last build; Be careful not 
# to have files in this folder that have same name as binary. Does a wildcard search: /filename*
clean:
	find ${ROOT_DIR} -type f -name '${BINARY}*' -exec rm -f {} \;

# none of our targets are files, so all are PHONY
.PHONY: default, build, all, build_all, clean