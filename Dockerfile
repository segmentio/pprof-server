FROM golang:1.9

RUN apt-get update -y && \
    apt-get install -y ca-certificates graphviz

COPY ./pprof-server /usr/local/bin/pprof-server

EXPOSE 6061

ENTRYPOINT ["pprof-server"]

