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

test:
	cd test; ./run

docker:
	@echo ${BRANCHNAME}
	@echo ${VERSION}
	docker build -t ${CONTAINER} .
	docker tag ${CONTAINER} ${CONTAINER-CI}
	docker tag ${CONTAINER} ${CONTAINER-LATEST}
	docker push ${CONTAINER-CI}
	docker push ${CONTAINER}
	docker push ${CONTAINER-LATEST}

package:
	rm -rf rpm
	rm -f replicator-*.rpm
	mkdir -p rpm/usr/bin
	mkdir -p rpm/usr/lib/systemd/system
	mkdir -p rpm/data/ubanquity/conf
	mkdir -p rpm/data/ubanquity/log
	cp replicator   							rpm/usr/bin/
	cp replicator.service   					rpm/usr/lib/systemd/system/
	cp conf/replicator_static.toml              rpm/data/ubanquity/conf
	cp conf/replicator_dynamic.toml             rpm/data/ubanquity/conf
	cp conf/dwh_map.yaml                        rpm/data/ubanquity/conf
	fpm -s dir \
	    -t rpm \
		-n replicator \
		-v 5 \
		--iteration 2 \
		--config-files /data/ubanquity/conf/replicator_static.toml \
		--config-files /data/ubanquity/conf/replicator_dynamic.toml  \
		--config-files /data/ubanquity/conf/dwh_map.yaml \
		--config-files /usr/lib/systemd/system/replicator.service \
		--rpm-attr 640,ubanquity,ubanquity:/data/ubanquity/conf/replicator_static.toml \
		--rpm-attr 640,ubanquity,ubanquity:/data/ubanquity/conf/replicator_dynamic.toml \
		--rpm-attr 640,ubanquity,ubanquity:/data/ubanquity/conf/dwh_map.yaml \
		--rpm-attr 750,ubanquity,ubanquity:/data/ubanquity/log \
		--rpm-attr 750,ubanquity,ubanquity:/data/ubanquity/conf \
		--license Proprietary \
		--vendor ubanquity \
		-m it@ubanquity.com \
		--description "Ubanquity Postgres Replicator" \
		--url "https://ubanquity.com"  \
		-C rpm .

install:
	sudo mv replicator-*.rpm /data/pkgrepo
	sudo /usr/local/bin/repo

clean:
	rm -f ${BINARY}

.PHONY: test

