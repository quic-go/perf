# quic-go perf

This is a (partial) implementation of the [QUIC Perf Protocol](https://datatracker.ietf.org/doc/html/draft-banks-quic-performance-00).

## Usage

Choose a command: `server`, `throughput`, or `handshake`. Use `--help` after a command to see its options.

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

### Handshakes per second
```commandline
go run cmd/main.go handshake --address=<server ip>:<server port> --concurrency=64 --duration=30s
```

The client repeatedly connects and closes without resumption or application data.
Defaults match MsQuic's HPS preset: 16 concurrent handshakes per CPU for 12 seconds.
JSON reports successful, failed, and incomplete handshakes, plus HPS. Only client
completions before the deadline count; setup and final cleanup are excluded.
A run with no successful handshakes exits with an error.

The server uses an RSA-2048 / SHA-256 certificate, matching MsQuic's OpenSSL helper.
TLS defaults still differ; align the key exchange and cipher suite before comparing
cryptographic workloads. Disable key logging and qlog for performance measurements.
