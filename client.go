package perf

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"

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
	return handleClientStream(str, uploadBytes, downloadBytes)
}

func handleClientStream(str io.ReadWriteCloser, uploadBytes, downloadBytes uint64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, downloadBytes)
	if _, err := str.Write(b); err != nil {
		return err
	}

	// upload data
	b = make([]byte, 16*1024)
	for uploadBytes > 0 {
		if uploadBytes < uint64(len(b)) {
			b = b[:uploadBytes]
		}
		n, err := str.Write(b)
		if err != nil {
			return err
		}
		uploadBytes -= uint64(n)
	}
	if err := str.Close(); err != nil {
		return err
	}

	// download data
	b = b[:cap(b)]
	remaining := downloadBytes
	for remaining > 0 {
		n, err := str.Read(b)
		if uint64(n) > remaining {
			return fmt.Errorf("server sent more data than expected, expected %d, got %d", downloadBytes, remaining+uint64(n))
		}
		remaining -= uint64(n)
		if err != nil {
			if err == io.EOF {
				if remaining == 0 {
					break
				}
				return fmt.Errorf("server didn't send enough data, expected %d, got %d", downloadBytes, downloadBytes-remaining)
			}
			return err
		}
	}
	return nil
}
