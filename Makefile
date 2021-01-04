sources := $(wildcard *.go) $(wildcard ./cmd/pprof-server/*.go)
branch ?= $(shell git rev-parse --abbrev-ref HEAD)
commit ?= $(shell git rev-parse --short=7 HEAD)
go := GOARGS=-mod=vendor go

all: test pprof-server

test: vendor
	$(go) test ./...

clean:
	rm -f pprof-server

vendor: ./vendor/modules.txt

./vendor/modules.txt: go.mod go.sum
	go mod vendor

pprof-server: vendor $(sources)
	$(go) build ./cmd/pprof-server

docker.version ?= $(subst /,-,$(branch))-$(commit)
docker.image ?= "528451384384.dkr.ecr.us-west-2.amazonaws.com/pprof-server:$(docker.version)"
docker: vendor
	docker build -t $(docker.image) -f Dockerfile .

publish: docker
	docker push $(docker.image)

.PHONY: all test clean vendor docker publish
