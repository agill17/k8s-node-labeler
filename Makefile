VERSION ?= 0.1.0
IMG ?= agill17/k8s-node-labeler
.DEFAULT_GOAL := all

build:
	docker build . -t ${IMG}:${VERSION}

push:
	docker push ${IMG}:${VERSION}

all: build push
