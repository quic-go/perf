package perf

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/require"
)

func TestClientScenarios(t *testing.T) {
	tlsConf, err := generateSelfSignedTLSConfig()
	require.NoError(t, err)
	tlsConf.NextProtos = []string{ALPN}
	ln, err := quic.ListenAddr("127.0.0.1:0", tlsConf, config)
	require.NoError(t, err)
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept(context.Background())
			if err != nil {
				return
			}
			go handleConn(conn)
		}
	}()

	measurements := make(chan any, 1)
	require.NoError(t, RunClient(ln.Addr().String(), 12345, 54321, nil, measurements))
	throughput := (<-measurements).(Result)
	require.Equal(t, "final", throughput.Type)
	require.Equal(t, uint64(12345), throughput.UploadBytes)
	require.Equal(t, uint64(54321), throughput.DownloadBytes)

	require.NoError(t, RunHandshakeClient(ln.Addr().String(), 4, time.Second, nil, measurements))
	result := (<-measurements).(HandshakeResult)
	require.Equal(t, "final", result.Type)
	require.Greater(t, result.Handshakes, uint64(4))
	require.Zero(t, result.FailedHandshakes)
	require.LessOrEqual(t, result.IncompleteHandshakes, uint64(4))
	require.Equal(t, float64(result.Handshakes)/result.TimeSeconds, result.HandshakesPerSecond)

	// A silent UDP peer leaves every attempt unfinished until the run ends.
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	require.NoError(t, err)
	defer udpConn.Close()
	start := time.Now()
	err = RunHandshakeClient(udpConn.LocalAddr().String(), 2, 100*time.Millisecond, nil, measurements)
	require.Less(t, time.Since(start), 3*time.Second)
	require.ErrorContains(t, err, "no successful handshakes")
	result = (<-measurements).(HandshakeResult)
	require.Zero(t, result.Handshakes)
	require.Zero(t, result.FailedHandshakes)
	require.Equal(t, uint64(2), result.IncompleteHandshakes)
}
