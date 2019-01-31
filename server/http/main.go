package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

var port string

func init() {
	flag.StringVar(&port, "port", "8080", "The port on which to serve")
}

// handlePing writes a pong back to the HTTP response.
func handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("http pong"))
}

func main() {
	flag.Parse()

	http.HandleFunc("/ping", handlePing)

	fmt.Println("Serving on port:", port)
	log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), nil))
}
