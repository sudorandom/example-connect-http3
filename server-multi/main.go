package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"

	"connectrpc.com/connect"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	elizav1 "buf.build/gen/go/connectrpc/eliza/protocolbuffers/go/connectrpc/eliza/v1"
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

func main() {
	mux := http.NewServeMux()
	mux.Handle(elizav1connect.NewElizaServiceHandler(&server{}))

	eg, ctx := errgroup.WithContext(context.Background())

	addr := "127.0.0.1:6660"
	log.Printf("Starting connectrpc on %s", addr)
	h3srv := http3.Server{
		Addr:    addr,
		Handler: mux,
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
	eg.Go(func() error {
		return srv.ListenAndServeTLS("cert.crt", "cert.key")
	})
	if err := eg.Wait(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
