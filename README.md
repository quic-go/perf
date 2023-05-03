# quic-go perf

This is a (partial) implementation of the [QUIC Perf Protocol](https://datatracker.ietf.org/doc/html/draft-banks-quic-performance-00).

## Usage

### Server
```commandline
go run cmd/main.go -run-server=true -server-address=0.0.0.0:<server port>
```

A pprof endpoint is available at port 6060.

### Client
```commandline
go run cmd/main.go -server-address=<server ip>:<server port> -upload-bytes=<N> -download-bytes=<M>
```

A pprof endpoint is available at port 6061.
