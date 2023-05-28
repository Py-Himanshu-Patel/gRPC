package main

import (
	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	pb "server/ecommerce"
)

var (
	port    = ":50051"
	crtFile = "cert/server.crt"
	keyFile = "cert/server.key"
)

type server struct {
	pb.UnimplementedProductInfoServer
	productMap map[string]*pb.Product
}

func main() {
	// Read and parse a public/private key pair and create
	// a certificate to enable TLS.
	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load key pair: %s", err)
	}
	// Enable TLS for all incoming connections by adding
	// certificates as TLS server credentials.
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}
	// Create a new gRPC server instance by passing TLS server credentials.
	s := grpc.NewServer(opts...)

	// Register the implemented service to the newly created
	// gRPC server by calling generated APIs.
	pb.RegisterProductInfoServer(s, &server{})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Bind the gRPC server to the listener and start listening
	// to incoming messages on the port (50051)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
