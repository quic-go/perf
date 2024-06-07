package main

import (
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/quic-go/perf"
)

type Options struct {
	RunServer     bool   `long:"run-server" description:"run as server, default: false"`
	TLS           bool   `long:"tcp-tls" description:"run on TLS 1.3/TCP, default: false"`
	KeyLogFile    string `long:"key-log" description:"export TLS keys"`
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

	var keyLogFile io.Writer
	if opt.KeyLogFile != "" {
		f, err := os.Create(opt.KeyLogFile)
		if err != nil {
			log.Fatalf("failed to create key log file: %s", err)
		}
		defer f.Close()
		keyLogFile = f
	}

	if opt.RunServer {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
		if opt.TLS {
			if err := perf.RunTLSServer(opt.ServerAddress, keyLogFile); err != nil {
				log.Fatal(err)
			}
			return
		}
		if err := perf.RunQUICServer(opt.ServerAddress, keyLogFile); err != nil {
			log.Fatal(err)
		}
	} else {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6061", nil))
		}()
		if opt.TLS {
			if err := perf.RunTLSClient(
				opt.ServerAddress,
				perf.ParseBytes(opt.UploadBytes),
				perf.ParseBytes(opt.DownloadBytes),
				keyLogFile,
			); err != nil {
				log.Fatal(err)
			}
			return
		}
		if err := perf.RunQUICClient(
			opt.ServerAddress,
			perf.ParseBytes(opt.UploadBytes),
			perf.ParseBytes(opt.DownloadBytes),
			keyLogFile,
		); err != nil {
			log.Fatal(err)
		}
	}
}
