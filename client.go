package perf

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

func RunClient(addr string, uploadBytes, downloadBytes uint64) error {
	conn, err := quic.DialAddr(
		addr,
		&tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{ALPN},
		},
		nil,
	)
	if err != nil {
		return err
	}
	str, err := conn.OpenStream()
	if err != nil {
		return err
	}
	uploadTook, downloadTook, err := handleClientStream(str, uploadBytes, downloadBytes)
	if err != nil {
		return err
	}
	log.Printf("uploaded %s: %.2fs (%s/s)", formatBytes(uploadBytes), uploadTook.Seconds(), formatBytes(bandwidth(uploadBytes, uploadTook)))
	log.Printf("downloaded %s: %.2fs (%s/s)", formatBytes(downloadBytes), downloadTook.Seconds(), formatBytes(bandwidth(downloadBytes, downloadTook)))
	fmt.Printf("{ \"latencies\": [ %f ]}\n", uploadTook.Seconds() + downloadTook.Seconds())
	return nil
}

func handleClientStream(str io.ReadWriteCloser, uploadBytes, downloadBytes uint64) (uploadTook, downloadTook time.Duration, err error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, downloadBytes)
	if _, err := str.Write(b); err != nil {
		return 0, 0, err
	}

	// upload data
	b = make([]byte, 16*1024)
	uploadStart := time.Now()
	for uploadBytes > 0 {
		if uploadBytes < uint64(len(b)) {
			b = b[:uploadBytes]
		}
		n, err := str.Write(b)
		if err != nil {
			return 0, 0, err
		}
		uploadBytes -= uint64(n)
	}
	if err := str.Close(); err != nil {
		return 0, 0, err
	}
	uploadTook = time.Since(uploadStart)

	// download data
	b = b[:cap(b)]
	remaining := downloadBytes
	downloadStart := time.Now()
	for remaining > 0 {
		n, err := str.Read(b)
		if uint64(n) > remaining {
			return 0, 0, fmt.Errorf("server sent more data than expected, expected %d, got %d", downloadBytes, remaining+uint64(n))
		}
		remaining -= uint64(n)
		if err != nil {
			if err == io.EOF {
				if remaining == 0 {
					break
				}
				return 0, 0, fmt.Errorf("server didn't send enough data, expected %d, got %d", downloadBytes, downloadBytes-remaining)
			}
			return 0, 0, err
		}
	}
	return uploadTook, time.Since(downloadStart), nil
}
