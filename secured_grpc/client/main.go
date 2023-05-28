package main

import (
	// pb "client/ecommerce"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

var (
	address  = "localhost:50051"
	hostname = "localhost"
	crtFile  = "cert/server.crt"
)

func main() {
	// Read and parse a public certificate and create a certificate to enable TLS.
	creds, err := credentials.NewClientTLSFromFile(crtFile, hostname)
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// c := pb.NewProductInfoClient(conn)
	// Skip RPC method invocation.
}
