package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/quic-go/quic-go/http3"
)

const (
	skipVerify = true
	url        = "https://127.0.0.1:6660/connectrpc.eliza.v1.ElizaService/Say"

	// skipVerify = false
	// url        = "https://demo.connectrpc.com/connectrpc.eliza.v1.ElizaService/Say"

	reqBody = `{"sentence": "Hello World!"}`
)

func main() {
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipVerify,
		},
	}
	defer roundTripper.Close()
	client := &http.Client{
		Transport: roundTripper,
	}

	log.Println("connect: ", url)
	log.Println("send: ", reqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	defer resp.Body.Close()

	log.Println("recv: ", string(body))
}
