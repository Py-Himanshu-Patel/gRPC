package main

import (
	"context"
	"log"
	"time"

	// contains the generated code we created from the protobuf compiler
	pb "productinfo/client/ecommerce"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection with the server from the
	// provided address (“localhost: 50051”)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// Pass the connection and create a stub. This stub
	// instance contains all the remote methods to invoke the server.
	c := pb.NewProductInfoClient(conn)

	name := "Apple iPhone 11"
	description := `Meet Apple iPhone 11. All-new dual-camera 
	system with Ultra Wide and Night mode.`

	// Create a Context to pass with the remote call. Here
	// the Context object contains metadata such as the identity
	// of the end user, authorization tokens, and the request’s
	// deadline and it will exist during the lifetime of the request.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Call addProduct method with product details.
	// This returns a product ID if the action completed successfully.
	// Otherwise it returns an error.
	r, err := c.AddProduct(ctx, &pb.Product{Name: name, Description: description})
	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}
	log.Printf("Product ID: %s added successfully", r.Value)

	// Call getProduct with the product ID
	product, err := c.GetProduct(ctx, &pb.ProductID{Value: r.Value})
	if err != nil {
		log.Fatalf("Could not get product: %v", err)
	}
	log.Printf("Product: %s", product.String())
}
