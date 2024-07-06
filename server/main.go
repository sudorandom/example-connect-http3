package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
	"connectrpc.com/connect"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

var _ elizav1connect.ElizaServiceHandler = (*server)(nil)

type server struct {
	elizav1connect.UnimplementedElizaServiceHandler
}

// Say implements elizav1connect.ElizaServiceHandler.
func (s *server) Say(ctx context.Context, req *connect.Request[elizav1.SayRequest]) (*connect.Response[elizav1.SayResponse], error) {
	slog.Info("Say()", "req", req)
	return connect.NewResponse(&elizav1.SayResponse{
		Sentence: req.Msg.GetSentence(),
	}), nil
}

func GenX509KeyPair() (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName:         "localhost",
			Country:            []string{"DK"},
			Organization:       []string{"squirrels"},
			OrganizationalUnit: []string{"SQ"},
		},
		NotBefore:   now,
		NotAfter:    now.AddDate(0, 0, 1), // Valid for one day
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	return outCert, nil
}

func main() {
	mux := http.NewServeMux()
	mux.Handle(elizav1connect.NewElizaServiceHandler(&server{}))

	cert, err := GenX509KeyPair()
	if err != nil {
		log.Fatal(err)
	}

	eg, ctx := errgroup.WithContext(context.Background())

	addr := "127.0.0.1:6660"
	log.Printf("Starting connectrpc on %s", addr)
	h3srv := http3.Server{
		Addr:    addr,
		Handler: mux,
		TLSConfig: http3.ConfigureTLSConfig(&tls.Config{
			Certificates: []tls.Certificate{cert},
		}),
		QUICConfig: &quic.Config{},
	}

	srv := http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	eg.Go(func() error {
		return h3srv.ListenAndServeTLS("cert.crt", "cert.key")
	})
	eg.Go(srv.ListenAndServe)
	if err := eg.Wait(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
