package main

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/quic-go/perf"
)

type Options struct {
	KeyLogFile string   `long:"key-log" description:"export TLS keys"`
	Address    string   `long:"address" required:"true" description:"address to listen on or connect to"`
	Server     struct{} `command:"server" description:"run the server"`
	Throughput struct {
		UploadBytes   string `long:"upload-bytes" description:"upload bytes #[KMG]"`
		DownloadBytes string `long:"download-bytes" description:"download bytes #[KMG]"`
	} `command:"throughput" description:"measure throughput"`
	Handshake struct {
		Concurrency int           `long:"concurrency" description:"concurrent handshakes (default: 16 per CPU)"`
		Duration    time.Duration `long:"duration" default:"12s" description:"handshake measurement duration"`
	} `command:"handshake" description:"measure handshakes per second"`
}

func main() {
	var opt Options
	opt.Handshake.Concurrency = 16 * runtime.NumCPU()
	parser := flags.NewParser(&opt, flags.Default)
	args, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := errors.AsType[*flags.Error](err); ok && flagsErr.Type == flags.ErrHelp {
			return
		}
		os.Exit(1)
	}
	if len(args) > 0 {
		log.Fatalf("unexpected arguments: %v", args)
	}
	if opt.Address == "" {
		log.Fatal("address must not be empty")
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

	pprofAddress := "0.0.0.0:6061"
	if parser.Active.Name == "server" {
		pprofAddress = "0.0.0.0:6060"
	}
	go func() {
		log.Println(http.ListenAndServe(pprofAddress, nil))
	}()

	measurements := make(chan any)
	go func() {
		defer close(measurements)
		switch parser.Active.Name {
		case "server":
			err = perf.RunServer(opt.Address, keyLogFile)
		case "throughput":
			err = perf.RunClient(
				opt.Address,
				perf.ParseBytes(opt.Throughput.UploadBytes),
				perf.ParseBytes(opt.Throughput.DownloadBytes),
				keyLogFile,
				measurements,
			)
		case "handshake":
			err = perf.RunHandshakeClient(opt.Address, opt.Handshake.Concurrency, opt.Handshake.Duration, keyLogFile, measurements)
		}
	}()
	encoder := jsontext.NewEncoder(os.Stdout)
	for measurement := range measurements {
		if err := json.MarshalEncode(encoder, measurement); err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
