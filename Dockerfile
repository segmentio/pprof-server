FROM 528451384384.dkr.ecr.us-west-2.amazonaws.com/segment-golang:1.17.6

RUN apk add --update --no-cache ca-certificates graphviz perl

WORKDIR $GOPATH/src/github.com/segmentio/pprof-server

COPY . .
COPY ./flamegraph.pl /usr/local/bin/flamegraph.pl
RUN go build -o /usr/local/bin/pprof-server ./cmd/pprof-server

EXPOSE 6061

ENTRYPOINT ["/usr/local/bin/pprof-server"]
