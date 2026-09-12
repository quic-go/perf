package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

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
}

func main() {
	var opt Options
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

	switch parser.Active.Name {
	case "server":
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
		err = perf.RunServer(opt.Address, keyLogFile)
	case "throughput":
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6061", nil))
		}()
		err = perf.RunClient(
			opt.Address,
			perf.ParseBytes(opt.Throughput.UploadBytes),
			perf.ParseBytes(opt.Throughput.DownloadBytes),
			keyLogFile,
		)
	}
	if err != nil {
		log.Fatal(err)
	}
}
