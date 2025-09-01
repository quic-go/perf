package perf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatBytes(t *testing.T) {
	require.Equal(t, "999 B", formatBytes(999))
	require.Equal(t, "1.00 KiB", formatBytes(1024))
	require.Equal(t, "1.00 MiB", formatBytes(1024*1024))
	require.Equal(t, "1.00 GiB", formatBytes(1024*1024*1024))
	require.Equal(t, "1.00 TiB", formatBytes(1024*1024*1024*1024))
	require.Equal(t, "3.45 MiB", formatBytes(1024*1024*345/100))
}

func TestFormatBandwidth(t *testing.T) {
	require.Equal(t, "800.00 bps", formatBandwidth(100, time.Second))
	require.Equal(t, "400.00 bps", formatBandwidth(100, 2*time.Second))
	require.Equal(t, "1.00 kbps", formatBandwidth(125, time.Second))
	require.Equal(t, "8.00 mbps", formatBandwidth(1e6, time.Second))
	require.Equal(t, "1.60 mbps", formatBandwidth(1e6, 5*time.Second))
	require.Equal(t, "9.87 kbps", formatBandwidth(1234, time.Second))
}
