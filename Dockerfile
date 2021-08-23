FROM golang:1.17.0

RUN apt-get update -y && \
    apt-get install -y ca-certificates graphviz

WORKDIR $GOPATH/src/github.com/segmentio/pprof-server
COPY . .
RUN cp ./flamegraph.pl /usr/local/bin/flamegraph.pl
RUN GOARGS=-mod=vendor make clean pprof-server \
    && mv pprof-server /usr/local/bin/pprof-server

EXPOSE 6061

ENTRYPOINT ["pprof-server"]
