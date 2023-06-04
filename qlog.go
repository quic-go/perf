package perf

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

func maybeAddQlogger(conf *quic.Config) {
	qlogDir := os.Getenv("QLOGDIR")
	if qlogDir == "" {
		return
	}
	if err := os.MkdirAll(qlogDir, 0660); err != nil {
		log.Fatalf("failed to create QLOGDIR (%s): %s", qlogDir, err)
	}
	if conf.Tracer != nil {
		log.Fatal("failed to add qlogger: quic.Config already has a tracer set")
	}
	conf.Tracer = func(ctx context.Context, perspective logging.Perspective, id quic.ConnectionID) logging.ConnectionTracer {
		role := "server"
		if perspective == logging.PerspectiveClient {
			role = "client"
		}
		filename := fmt.Sprintf("log_%x_%s.qlog", id.Bytes(), role)
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("failed to create qlog file %s: %s", filename, err)
		}
		log.Printf("created qlog file: %s", filename)
		return qlog.NewConnectionTracer(file, perspective, id)
	}
}
