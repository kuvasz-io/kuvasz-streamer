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

all: web check build vulncheck

web:
	cd web; yarn install; yarn build --outDir ../streamer/admin

check:
	staticcheck -checks=all ./...
	go vet ./...
	golangci-lint run
	govulncheck ./...

build:
	go build -o ${BINARY} -ldflags="${LDFLAGS}" ./streamer/*.go

vulncheck:
	govulncheck -mode=binary kuvasz-streamer

release:
	goreleaser release --clean --snapshot

rpmrepo:
	cp dist/*.rpm /var/www/caddy/rpm
	/var/www/caddy/rpm; createrepo_c -v /var/www/caddy/rpm

aptrepo:
	aptly repo add kuvasz dist/*.deb
	aptly publish update --passphrase="${GPG_PASSPHRASE}" --batch=true stable filesystem:caddy:

test:
	cd test; ./run

docs:
	cd docs; jekyll build
	rm -rf /var/www/caddy/streamer/*
	cp -r docs/_site/* /var/www/caddy/streamer

clean:
	rm -rf ${BINARY} streamer/admin web/dist dist

.PHONY: web check build vulncheck release rpmrepo aptrepo test docs clean

