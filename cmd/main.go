package main

import (
	"crypto/tls"
	"flag"
	"github.com/quic-go/perf"
	"log"
)

var tlsConf *tls.Config

func main() {
	server := flag.Bool("run-server", false, "Should run as server")
	serverAddr := flag.String("server-address", "", "Server address")
	uploadBytes := flag.Uint64("upload-bytes", 0, "Upload bytes")
	downloadBytes := flag.Uint64("download-bytes", 0, "Download bytes")
	flag.Parse()

	if *server {
		if err := perf.RunServer(*serverAddr); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := perf.RunClient(*serverAddr, *uploadBytes, *downloadBytes); err != nil {
			log.Fatal(err)
		}
	}
}
