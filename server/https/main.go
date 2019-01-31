package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var port, sslPort, cert, key string

func init() {
	flag.StringVar(&port, "port", "8080", "The port on which to serve non-SSL")
	flag.StringVar(&sslPort, "ssl-port", "443", "The port on which to serve SSL")
	flag.StringVar(&cert, "cert", "../ssl/server.crt", "Path to the SSL certificate for the server")
	flag.StringVar(&key, "key", "../ssl/server.key", "Path to the SSL private key for the server")
}

// handleUpgrade writes "ok" back to the HTTP response to indicate that this
// server supports SSL/TLS for it's main API resources.
func handleUpgrade(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
}

// handlePing writes a pong back to the HTTP response.
func handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("https pong"))
}

func main() {
	flag.Parse()

	// HTTP server to handle the upgrade indicator endpoint.
	mux := http.NewServeMux()
	mux.HandleFunc("/connupgrade", handleUpgrade)

	// The server is blocking so has to be in a goroutine as we're running
	// more than one server.
	go func() {
		fmt.Println("Serving on port:", port)
		log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), mux))
	}()

	// Main HTTPS server providing our ping endpoint.
	sslMux := http.NewServeMux()
	sslMux.HandleFunc("/ping", handlePing)

	crt, err := ioutil.ReadFile(cert)
	if err != nil {
		log.Fatalln("read cert:", err)
	}

	// Add the self-signed cert to the CA pool.
	crtPool := x509.NewCertPool()
	if ok := crtPool.AppendCertsFromPEM(crt); !ok {
		log.Fatalln("could not append certificate to the pool")
	}

	// Spin up a server using the cert pool with our new cert appended. Also
	// configure TLS to verify the certificate.
	svr := &http.Server{
		Addr:    net.JoinHostPort("", sslPort),
		Handler: sslMux,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  crtPool,
		},
	}

	fmt.Println("Serving SSL on port:", sslPort)
	log.Fatal(svr.ListenAndServeTLS(cert, key))
}
