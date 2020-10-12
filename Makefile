VERSION ?= 0.1.0
IMG ?= agill17/k8s-node-labeler

build:
	docker build . -t ${IMG}:${VERSION}

push:
	docker push -t ${IMG}:${VERSION}

all: build push
