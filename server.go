package perf

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

func RunServer(addr string, keyLogFile io.Writer) error {
	tlsConf, err := generateSelfSignedTLSConfig()
	if err != nil {
		log.Fatal(err)
	}
	tlsConf.NextProtos = []string{ALPN}
	tlsConf.KeyLogWriter = keyLogFile

	conf := config.Clone()
	ln, err := quic.ListenAddr(addr, tlsConf, conf)
	if err != nil {
		return err
	}
	log.Println("Listening on", ln.Addr())
	defer ln.Close()
	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			return fmt.Errorf("accept errored: %w", err)
		}
		go func(conn *quic.Conn) {
			if err := handleConn(conn); err != nil {
				log.Printf("handling conn from %s failed: %s", conn.RemoteAddr(), err)
			}
		}(conn)
	}
}

func handleConn(conn *quic.Conn) error {
	for {
		str, err := conn.AcceptStream(context.Background())
		if err != nil {
			return err
		}
		go func(str *quic.Stream) {
			if err := handleServerStream(str); err != nil {
				log.Printf("handling stream from %s failed: %s", conn.RemoteAddr(), err)
			}
		}(str)
	}
}

func handleServerStream(str io.ReadWriteCloser) error {
	b := make([]byte, 8)
	if _, err := io.ReadFull(str, b); err != nil {
		return err
	}
	amount := binary.BigEndian.Uint64(b)
	b = make([]byte, 16*1024)
	// receive data until the client sends a FIN
	for {
		if _, err := str.Read(b); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	// send as much data as the client requested
	for amount > 0 {
		if amount < uint64(len(b)) {
			b = b[:amount]
		}
		n, err := str.Write(b)
		if err != nil {
			return err
		}
		amount -= uint64(n)
	}
	return str.Close()
}

func generateSelfSignedTLSConfig() (*tls.Config, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, pubKey, privKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	b, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
