package perf

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

type HandshakeResult struct {
	Type                 string  `json:"type"`
	TimeSeconds          float64 `json:"timeSeconds"`
	Handshakes           uint64  `json:"handshakes"`
	FailedHandshakes     uint64  `json:"failedHandshakes"`
	IncompleteHandshakes uint64  `json:"incompleteHandshakes"`
	HandshakesPerSecond  float64 `json:"handshakesPerSecond"`
}

// RunHandshakeClient measures full client handshake completions in a fixed time
// window. Closing connections and reusing UDP sockets is part of the workload;
// DNS resolution, socket creation, and final cleanup are outside the window.
func RunHandshakeClient(addr string, concurrency int, duration time.Duration, keyLogFile io.Writer, measurements chan<- any) error {
	if concurrency <= 0 || duration <= 0 {
		return errors.New("handshake concurrency and duration must be greater than zero")
	}
	remoteAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	serverName, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
		NextProtos:         []string{ALPN},
		KeyLogWriter:       keyLogFile,
		// No session cache: every connection performs a full handshake.
	}

	transports := make([]quic.Transport, concurrency)
	for i := range transports {
		udpConn, err := net.ListenUDP("udp", nil)
		if err != nil {
			return err
		}
		defer udpConn.Close()
		transports[i] = quic.Transport{Conn: udpConn}
		defer transports[i].Close()
	}

	results := make([]HandshakeResult, concurrency)
	deadline := time.Now().Add(duration)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	var wg sync.WaitGroup
	for i := range concurrency {
		wg.Go(func() {
			for time.Now().Before(deadline) {
				conn, err := transports[i].Dial(ctx, remoteAddr, tlsConf, config)
				switch {
				case !time.Now().Before(deadline):
					results[i].IncompleteHandshakes++
				case err != nil:
					results[i].FailedHandshakes++
				default:
					results[i].Handshakes++
				}
				if conn != nil {
					conn.CloseWithError(0, "")
				}
			}
		})
	}
	wg.Wait()

	result := HandshakeResult{Type: "final", TimeSeconds: duration.Seconds()}
	for _, r := range results {
		result.Handshakes += r.Handshakes
		result.FailedHandshakes += r.FailedHandshakes
		result.IncompleteHandshakes += r.IncompleteHandshakes
	}
	result.HandshakesPerSecond = float64(result.Handshakes) / result.TimeSeconds
	measurements <- result
	if result.Handshakes == 0 {
		return errors.New("no successful handshakes")
	}
	return nil
}
