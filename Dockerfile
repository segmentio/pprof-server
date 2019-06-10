FROM golang:1.12.5

RUN apt-get update -y && \
    apt-get install -y ca-certificates graphviz

COPY ./pprof-server /usr/local/bin/pprof-server
COPY ./flamegraph.pl /usr/local/bin/flamegraph.pl

EXPOSE 6061

ENTRYPOINT ["pprof-server"]

