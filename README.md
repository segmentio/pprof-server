# pprof-server [![CircleCI](https://circleci.com/gh/segmentio/pprof-server.svg?style=shield)](https://circleci.com/gh/segmentio/pprof-server)
Web server exposing performance profiles of Go services.

## Building
```
govendor sync
```
```
go build ./cmd/pprof-server
```

## Running
```
./pprof-server -registry consul://localhost:8500
```
```
docker run -it --rm -p 6061:6061 segment/pprof-server -registry consul://172.17.0.1:8500
```

![Screenshot](./images/pprof-server.png)
