FROM golang:1.17.1-alpine3.14   
RUN apk add --update --no-cache \
           ca-certificates  make graphviz 

WORKDIR $GOPATH/src/github.com/segmentio/pprof-server
COPY . .
RUN cp ./flamegraph.pl /usr/local/bin/flamegraph.pl
RUN GOARGS=-mod=vendor make clean pprof-server \
    && mv pprof-server /usr/local/bin/pprof-server

EXPOSE 6061

ENTRYPOINT ["pprof-server"]
