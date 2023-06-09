package main

import (
	"context"
	"log"
	"os"
	"time"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	pb "grpc_prod/proto-gen"
)

//const (
//	address = "localhost:50051"
//)

func getHostName() string {
	hostname := os.Getenv("hostname")
	if hostname == "" {
		hostname = "localhost"
	}
	// add pre decided port to hostname
	hostname = hostname + ":50051"
	return hostname
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(getHostName(), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewProductInfoClient(conn)

	// Contact the server and print out its response.
	name := "Sumsung S10"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2019"
	price := float32(700.0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.AddProduct(ctx, &pb.Product{Name: name, Description: description, Price: price})
	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}
	log.Printf("Product ID: %s added successfully", r.Value)

	product, err := c.GetProduct(ctx, &wrapper.StringValue{Value: r.Value})
	if err != nil {
		log.Fatalf("Could not get product: %v", err)
	}
	log.Printf("Product: ", product.String())
}
