package perf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var kmgMap = map[string]uint64{"K": 1024, "M": 1024 * 1024, "G": 1024 * 1024 * 1024}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatBandwidth(bytes uint64, d time.Duration) string {
	b := float64(8*bytes) / d.Seconds()
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%.2f bps", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cbps", float64(b)/float64(div), "kmgtpe"[exp])
}

func ParseBytes(input string) uint64 {
	if input == "" {
		return 0
	}
	var kmg uint64 = 1
	for s, v := range kmgMap {
		if strings.ToUpper(input[len(input)-1:]) == s {
			input = input[:len(input)-1]
			kmg = v
			break
		}
	}
	num, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		panic("invalid kmg number")
	}
	return num * kmg
}
