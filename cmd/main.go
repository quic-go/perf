package main

import (
	"crypto/tls"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/quic-go/perf"
)

type Options struct {
	RunServer      bool    `long:"run-server" description:"run as server, default: false"`
	ServerAddress  string  `long:"server-address" description:"server address, required"`
	UploadBytes    uint64  `long:"upload-bytes" description:"upload bytes"`
	DownloadBytes  uint64  `long:"download-bytes" description:"download bytes"`
}

var tlsConf *tls.Config

func main() {
	var opt Options
	parser := flags.NewParser(&opt, flags.IgnoreUnknown)
	_, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	if opt.ServerAddress == "" {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	if opt.RunServer {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
		if err := perf.RunServer(opt.ServerAddress); err != nil {
			log.Fatal(err)
		}
	} else {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6061", nil))
		}()
		if err := perf.RunClient(opt.ServerAddress, opt.UploadBytes, opt.DownloadBytes); err != nil {
			log.Fatal(err)
		}
	}
}
