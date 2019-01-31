# Self Sign
A proof-of-concept set of Golang applications to illustrate self-signed
SSL certificates.

## Summary
Within this repo are 3 applications:
1. A standard HTTP web server
2. A HTTPS web server using a self-signed SSL certificate
3. A client

The client simple sends a "ping" to the server and expects to get a
"pong" back again.

The client is able to determine what protocol (http or https) to use when
making it's ping request to the server. A https connection is preferred
and http is used as a fallback. This determination is achieved by calling
a /connupgrade endpoint on the server.

If this endpoint returns 200 the client assumes the server supports HTTPS,
if it returns a 404 the client assumes the server only supports HTTP. The
client makes two ping request to illustrate that the protocol lookup only
happens once and is then cached for the lifetime of the client.

## Usage
Build the binaries:
```bash
$ make
```

### Client
```bash
Usage of ./bin/client-linux-amd64:
    -cert string
        Path to the SSL certificate for the server (default "ssl/server.crt")
    -port string
        The port on which to ping non-SSL (default "8080")
    -server string
        Server for the server to ping (default "localhost")
    -ssl-port string
        The port on which to ping SSL (default "443")
```

### Server (HTTP)
```bash
Usage of ./bin/server-http-linux-amd64:
    -port string
        The port on which to serve (default "8080")
```

### Server (HTTPS)
```bash
Usage of ./bin/server-https-linux-amd64:
    -cert string
        Path to the SSL certificate for the server (default "../ssl/server.crt")
    -key string
        Path to the SSL private key for the server (default "../ssl/server.key")
    -port string
        The port on which to serve non-SSL (default "8080")
    -ssl-port string
        The port on which to serve SSL (default "443")
```
