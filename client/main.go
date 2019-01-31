package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var (
	cert, server, port, sslPort string
	protocolCache               = make(map[string]protocol)
)

func init() {
	flag.StringVar(&cert, "cert", "ssl/server.crt", "Path to the SSL certificate for the server")
	flag.StringVar(&server, "server", "localhost", "Server for the server to ping")
	flag.StringVar(&port, "port", "8080", "The port on which to ping non-SSL")
	flag.StringVar(&sslPort, "ssl-port", "443", "The port on which to ping SSL")
}

type protocol string

const (
	protocolHTTP  protocol = "http"
	protocolHTTPS protocol = "https"
)

func main() {
	flag.Parse()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Using this is pretty poor. In an ideal world you'd have
				// a certificate on the remote server that is properly signed
				// by a root CA.
				InsecureSkipVerify: true,
			},
		},
	}

	fmt.Println("Ping one")
	res, err := request(client, server, "/ping", port, sslPort, nil)
	if err != nil {
		log.Fatalln("do request:", err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("read response:", err)
	}

	fmt.Println(string(b))

	fmt.Println()
	fmt.Println("Ping two")
	res, err = request(client, server, "/ping", port, sslPort, nil)
	if err != nil {
		log.Fatalln("do request:", err)
	}

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("read response:", err)
	}

	fmt.Println(string(b))
}

// request sends an HTTP request and returns an HTTP response for the given
// uri and server.
//
// The function will detect from the remote server if the request can be made
// using HTTPS or, as a fallback, standard HTTP.
func request(client *http.Client, server, uri, port, sslPort string, body io.Reader) (*http.Response, error) {
	pc, err := svrProtocol(client, server, port)
	if err != nil {
		return nil, err
	}

	remote := net.JoinHostPort(server, sslPort)
	if pc == protocolHTTP {
		remote = net.JoinHostPort(server, port)
	}

	req, err := http.NewRequest(http.MethodGet, string(pc)+"://"+remote+uri, body)
	if err != nil {
		log.Fatalln("build request:", err)
	}

	return client.Do(req)
}

// svrProtocol returns the supported protocol to use when communicating with
// the given server.
//
// The protocol is determined by the existence of a working /connupgrade
// endpoint being served via HTTP on the given port. If this endpoint returns
// a 200 response we assume the server is serving it's main resources on an
// SSL/TLS server. If the endpoint returns a 404 we assume the server only
// supports a HTTP server for it's resources.
func svrProtocol(client *http.Client, server, port string) (protocol, error) {
	if p, ok := protocolCache[server]; ok {
		return p, nil
	}

	log.Printf("no cached protocol for server %q: attempting HTTPS upgrade\n", server)

	req, err := http.NewRequest(http.MethodGet, "http://"+server+":"+port+"/connupgrade", nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var p protocol
	switch res.StatusCode {
	case http.StatusNotFound:
		p = protocolHTTP
	case http.StatusOK:
		p = protocolHTTPS
	default:
		return "", fmt.Errorf("unexpected HTTP status code: %d", res.StatusCode)
	}

	protocolCache[server] = p
	return p, nil
}
