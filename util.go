package perf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var kmgMap = map[string]uint64{"K": 1024, "M": 1024 * 1024, "G": 1024 * 1024 * 1024}

func bandwidth(b uint64, d time.Duration) uint64 {
	return uint64(float64(b) / d.Seconds())
}

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
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func ToBytes(input string) (output uint64) {
	if len(input) == 0 {
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
