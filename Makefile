BINARY             := kuvasz-streamer
GITBRANCH          := $(shell git branch | grep \* | cut -d ' ' -f2)
CI_COMMIT_REF_NAME ?= ${GITBRANCH}
HASH               := $(shell git rev-parse --short HEAD)
COUNTREF           := $(shell git rev-list HEAD | wc -l | tr -d ' ')
VERSION            := ${CI_COMMIT_REF_NAME}-${COUNTREF}-${HASH}
BUILD              := $(shell date +%Y%m%d%H%M%S)
CONTAINER          := ${REGISTRY}/${BINARY}:${VERSION}
CONTAINER-CI       := ${REGISTRY}/${BINARY}:ci
CONTAINER-LATEST   := ${REGISTRY}/${BINARY}:${CI_COMMIT_REF_NAME}
LDFLAGS            += -X ${BINARY}.Version=${VERSION}
LDFLAGS            += -X ${BINARY}.Build=${BUILD}

all: build

check: 
	staticcheck -checks=all ./...
	go vet ./...
	golangci-lint run

build:
	go build -o ${BINARY} -ldflags="${LDFLAGS}" ./cmd/*.go

release:
	goreleaser release --clean --snapshot

test:
	cd test; ./run

docs:
	cd docs; jekyll build
	rm -rf /var/www/caddy/streamer/*
	cp -r docs/_site/* /var/www/caddy/streamer

docker:
	@echo ${BRANCHNAME}
	@echo ${VERSION}
	docker build -t ${CONTAINER} .
	docker tag ${CONTAINER} ${CONTAINER-CI}
	docker tag ${CONTAINER} ${CONTAINER-LATEST}
	docker push ${CONTAINER-CI}
	docker push ${CONTAINER}
	docker push ${CONTAINER-LATEST}

clean:
	rm -f ${BINARY}

.PHONY: check build test docs package install clean

