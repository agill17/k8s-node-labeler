VERSION ?= 0.1.0
IMG ?= agill17/k8s-node-labeler

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build:
	docker build . -t $IMAG:$VERSION

push:
	docker push -t $IMG:$VERSION

all: build push