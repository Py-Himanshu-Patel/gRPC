package main

import (
	// pb "client/ecommerce"
	pb "client/ecommerce"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	// Contact the server and print out its response.
	name := "Sumsung S10"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2019"
	price := float32(700.0)
	// make call to server with a timeout of a second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Add a product
	r, err := c.AddProduct(ctx, &pb.Product{Name: name, Description: description, Price: price})
	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}
	log.Printf("Product ID: %s added successfully", r.Value)
	// get the same product details which we added
	product, err := c.GetProduct(ctx, &pb.ProductID{Value: r.Value})
	if err != nil {
		log.Fatalf("Could not get product: %v", err)
	}
	log.Printf("Product: %s", product.String())
}
