package main

import (
	// pb "client/ecommerce"
	pb "client/ecommerce"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
)

var (
	address  = "localhost:50051"
	hostname = "localhost"
	crtFile  = "cert/client.crt"
	keyFile  = "cert/client.key"
	caFile   = "cert/ca.crt"
)

func main() {
	// Create X.509 key pairs directly from the server certificate and key.
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}
	// Create a certificate pool from the CA.
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}
	// Append the client certificates from the CA to the certificate pool.
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:   hostname, // NOTE: this is required!
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProductInfoClient(conn)
	fmt.Println("Connection Established : ", c)
	// Skip RPC method invocation.
}
