package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/quic-go/perf"
)

type Options struct {
	RunServer     bool   `long:"run-server" description:"run as server, default: false"`
	ServerAddress string `long:"server-address" description:"server address, required"`
	UploadBytes   string `long:"upload-bytes" description:"upload bytes #[KMG]"`
	DownloadBytes string `long:"download-bytes" description:"download bytes #[KMG]"`
}

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
		if err := perf.RunClient(opt.ServerAddress, perf.ParseBytes(opt.UploadBytes), perf.ParseBytes(opt.DownloadBytes)); err != nil {
			log.Fatal(err)
		}
	}
}
