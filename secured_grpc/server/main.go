package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"net"
	pb "server/ecommerce"
)

var (
	port    = ":50051"
	crtFile = "cert/server.crt"
	keyFile = "cert/server.key"
	caFile  = "cert/ca.crt"
)

type server struct {
	pb.UnimplementedProductInfoServer
	productMap map[string]*pb.Product
}

// AddProduct implements ecommerce.AddProduct
func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	return &pb.ProductID{Value: in.Id}, nil
}

// GetProduct implements ecommerce.GetProduct
func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	value, exists := s.productMap[in.Value]
	if exists {
		return value, nil
	}
	return nil, errors.New("Product does not exist for the ID" + in.Value)
}

func main() {
	// Read and parse a public/private key pair and create
	// a certificate to enable TLS.
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load key pair: %s", err)
	}

	// Create a certificate pool from the CA.
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA to the certificate pool.
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certificate")
	}

	// Enable TLS for all incoming connections by creating TLS credentials.
	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		grpc.Creds(
			credentials.NewTLS(&tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{certificate},
				ClientCAs:    certPool,
			},
			)),
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
