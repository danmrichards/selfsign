package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var (
	certFile, keyFile, server, port, sslPort string
	protocolCache                            = make(map[string]protocol)
)

func init() {
	flag.StringVar(&certFile, "cert", "ssl/server.crt", "Path to the SSL certificate file for the server")
	flag.StringVar(&keyFile, "key", "ssl/server.key", "Path to the SSL private key file for the server")
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

	crt, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Fatal(err)
	}

	// Get the certificate pool.
	crtPool, err := x509.SystemCertPool()
	if err != nil {
		// Just log the error instead of a fatal. In many cases (e.g. when
		// running on windows) we won't be able to get the system cert pool
		// at all. Better to just log and attempt to use the default pool.
		log.Println("could not load system cert pool:", err)
	}
	if crtPool == nil {
		crtPool = x509.NewCertPool()
	}

	// Add the self-signed cert to the CA pool.
	if ok := crtPool.AppendCertsFromPEM(crt); !ok {
		log.Fatalln("could not append certificate to the pool")
	}

	xkp, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Spin up a HTTP client using the extended root CA pool and the key/pair
	// we generated above. The key pair will be presented to the other side of
	// the connection and verified. We could avoid this if we had a certificate
	// from a root CA already trusted by the client and the server. Or if you
	// want to live dangerously by disabling verification.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{xkp},
				RootCAs:      crtPool,
			},
		},
	}

	fmt.Println("ping one")
	res, err := request(client, server, "/ping", port, sslPort, nil)
	if err != nil {
		log.Fatalln("ping one request:", err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("ping one response:", err)
	}

	fmt.Println(string(b))

	fmt.Println()
	fmt.Println("ping two")
	res, err = request(client, server, "/ping", port, sslPort, nil)
	if err != nil {
		log.Fatalln("ping two request:", err)
	}

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("ping two response:", err)
	}

	fmt.Println(string(b))

	fmt.Println()
	fmt.Println("ping google just to prove the root certs still work")

	res, err = client.Get("https://www.google.com")
	if err != nil {
		log.Fatalln("google request:", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Fatalln("could not ping google")
	}
	fmt.Println("pinged google")
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
