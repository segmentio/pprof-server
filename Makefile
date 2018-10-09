DOCKER_REPO ?= segment/pprof-server:latest

sources := $(wildcard *.go) $(wildcard ./cmd/pprof-server/*.go)

all: test pprof-server

test:
	go test ./...

pprof-server: $(sources)
	go build ./cmd/pprof-server

docker-image: pprof-server
	docker build -t pprof-server -f Dockerfile .

publish-image: docker-image
	docker tag pprof-server ${DOCKER_REPO}
	docker push ${DOCKER_REPO}
