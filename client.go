package perf

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

type Result struct {
	Type          string  `json:"type"`
	TimeSeconds   float64 `json:"timeSeconds"`
	UploadBytes   uint64  `json:"uploadBytes"`
	DownloadBytes uint64  `json:"downloadBytes"`
}

const handshakeTimeout = 5 * time.Second

func RunQUICClient(addr string, uploadBytes, downloadBytes uint64, keyLogFile io.Writer) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()
	conn, err := quic.DialAddr(
		ctx,
		addr,
		&tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{ALPN},
			KeyLogWriter:       keyLogFile,
			MinVersion:         tls.VersionTLS13,
		},
		config,
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
	printResults(uploadBytes, downloadBytes, uploadTook, downloadTook, time.Since(start))
	return nil
}

func RunTLSClient(addr string, uploadBytes, downloadBytes uint64, keyLogFile io.Writer) error {
	start := time.Now()
	dialer := &tls.Dialer{
		Config: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{ALPN},
			KeyLogWriter:       keyLogFile,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	uploadTook, downloadTook, err := handleClientStream(&writeCloseReadWriteCloser{Conn: conn.(*tls.Conn)}, uploadBytes, downloadBytes)
	if err != nil {
		return err
	}
	printResults(uploadBytes, downloadBytes, uploadTook, downloadTook, time.Since(start))
	return nil
}

type writeCloseReadWriteCloser struct {
	Conn interface {
		io.ReadWriter
		CloseWrite() error
	}
}

var _ io.ReadWriteCloser = &writeCloseReadWriteCloser{}

func (c *writeCloseReadWriteCloser) Read(b []byte) (int, error)  { return c.Conn.Read(b) }
func (c *writeCloseReadWriteCloser) Write(b []byte) (int, error) { return c.Conn.Write(b) }
func (c *writeCloseReadWriteCloser) Close() error                { return c.Conn.CloseWrite() }

func handleClientStream(str io.ReadWriteCloser, uploadBytes, downloadBytes uint64) (uploadTook, downloadTook time.Duration, err error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, downloadBytes)
	if _, err := str.Write(b); err != nil {
		return 0, 0, err
	}

	// upload data
	b = make([]byte, 16*1024)
	uploadStart := time.Now()

	lastReportTime := time.Now()
	lastReportWrite := uint64(0)

	for uploadBytes > 0 {
		now := time.Now()
		if now.Sub(lastReportTime) >= time.Second {
			jsonB, err := json.Marshal(Result{
				TimeSeconds: now.Sub(lastReportTime).Seconds(),
				UploadBytes: lastReportWrite,
				Type:        "intermediary",
			})
			if err != nil {
				log.Fatalf("failed to marshal perf result: %s", err)
			}
			fmt.Println(string(jsonB))

			lastReportTime = now
			lastReportWrite = 0
		}

		if uploadBytes < uint64(len(b)) {
			b = b[:uploadBytes]
		}
		n, err := str.Write(b)
		if err != nil {
			return 0, 0, err
		}
		uploadBytes -= uint64(n)
		lastReportWrite += uint64(n)
	}

	if err := str.Close(); err != nil {
		return 0, 0, err
	}
	uploadTook = time.Since(uploadStart)

	// download data
	b = b[:cap(b)]
	remaining := downloadBytes
	downloadStart := time.Now()

	lastReportTime = time.Now()
	lastReportRead := uint64(0)

	for remaining > 0 {
		now := time.Now()
		if now.Sub(lastReportTime) >= time.Second {
			jsonB, err := json.Marshal(Result{
				TimeSeconds:   now.Sub(lastReportTime).Seconds(),
				DownloadBytes: lastReportRead,
				Type:          "intermediary",
			})
			if err != nil {
				log.Fatalf("failed to marshal perf result: %s", err)
			}
			fmt.Println(string(jsonB))

			lastReportTime = now
			lastReportRead = 0
		}

		n, err := str.Read(b)
		if uint64(n) > remaining {
			return 0, 0, fmt.Errorf("server sent more data than expected, expected %d, got %d", downloadBytes, remaining+uint64(n))
		}
		remaining -= uint64(n)
		lastReportRead += uint64(n)
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

func printResults(uploadBytes, downloadBytes uint64, uploadTook, downloadTook, total time.Duration) {
	log.Printf("uploaded %s: %.2fs (%s/s)", formatBytes(uploadBytes), uploadTook.Seconds(), formatBytes(bandwidth(uploadBytes, uploadTook)))
	log.Printf("downloaded %s: %.2fs (%s/s)", formatBytes(downloadBytes), downloadTook.Seconds(), formatBytes(bandwidth(downloadBytes, downloadTook)))
	json, err := json.Marshal(Result{
		TimeSeconds:   total.Seconds(),
		Type:          "final",
		UploadBytes:   uploadBytes,
		DownloadBytes: downloadBytes,
	})
	if err != nil {
		log.Fatalf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(json))
}
