# quic-go perf

This is a (partial) implementation of the [QUIC Perf Protocol](https://datatracker.ietf.org/doc/html/draft-banks-quic-performance-00).

## Usage

Choose a command: `server` or `throughput`. Use `--help` after a command to see its options.

### Server
```commandline
go run cmd/main.go server --address=0.0.0.0:<server port>
```

A pprof endpoint is available at port 6060.

### Throughput client
```commandline
go run cmd/main.go throughput --address=<server ip>:<server port> --upload-bytes=<N> --download-bytes=<M>
```

A pprof endpoint is available at port 6061.
