package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/quic-go/perf"
)

var tlsConf *tls.Config

func main() {
	server := flag.Bool("run-server", false, "Should run as server")
	serverAddr := flag.String("server-address", "", "Server address")
	uploadBytes := flag.Uint64("upload-bytes", 0, "Upload bytes")
	downloadBytes := flag.Uint64("download-bytes", 0, "Download bytes")
	flag.Parse()

	if *server {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
		if err := perf.RunServer(*serverAddr); err != nil {
			log.Fatal(err)
		}
	} else {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6061", nil))
		}()
		if err := perf.RunClient(*serverAddr, *uploadBytes, *downloadBytes); err != nil {
			log.Fatal(err)
		}
	}
}
