package main

import (
	"flag"
	"fmt"
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

	svr := http.NewServeMux()
	svr.HandleFunc("/connupgrade", handleUpgrade)

	go func() {
		fmt.Println("Serving on port:", port)
		log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), svr))
	}()

	svrSSL := http.NewServeMux()
	svrSSL.HandleFunc("/ping", handlePing)

	fmt.Println("Serving SSL on port:", sslPort)
	log.Fatal(http.ListenAndServeTLS(net.JoinHostPort("", sslPort), cert, key, svrSSL))
}
