package perf

import (
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/qlog"
)

const ALPN = "perf"

var config = &quic.Config{
	// use massive flow control windows here to make sure that flow control is not the limiting factor
	InitialStreamReceiveWindow:     1 << 20,
	InitialConnectionReceiveWindow: 1 << 20,
	MaxConnectionReceiveWindow:     1 << 30,
	MaxStreamReceiveWindow:         1 << 30,
	Tracer:                         qlog.DefaultConnectionTracer,
}
