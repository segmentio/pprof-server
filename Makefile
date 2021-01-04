sources := $(wildcard *.go) $(wildcard ./cmd/pprof-server/*.go)
branch ?= $(shell git rev-parse --abbrev-ref HEAD)
commit ?= $(shell git rev-parse --short=7 HEAD)

all: test pprof-server

test:
	go test ./...

clean:
	rm -f pprof-server

vendor:
	go mod vendor

pprof-server: $(sources)
	go build ./cmd/pprof-server

docker.version ?= $(subst /,-,$(branch))-$(commit)
docker.image ?= "528451384384.dkr.ecr.us-west-2.amazonaws.com/centrifuge:$(docker.version)"
docker: vendor
	docker build -t $(docker.image) -f Dockerfile .

publish: docker
	docker push $(docker.image)
