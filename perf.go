package perf

import "github.com/quic-go/quic-go"

const ALPN = "perf"

var config = &quic.Config{
	// use massive flow control windows here to make sure that flow control is not the limiting factor
	MaxConnectionReceiveWindow: 1 << 30,
	MaxStreamReceiveWindow:     1 << 30,
}
