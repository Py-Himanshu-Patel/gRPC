package main

import (
	"context"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	pb "grpc_prod/proto-gen"
	"grpc_prod/tracer"
	"log"
	"os"
	"time"
)

const (
	port = ":50051"
)

func getHostName() string {
	hostname := os.Getenv("hostname")
	if hostname == "" {
		hostname = "localhost"
	}
	// add pre decided port to hostname
	hostname = hostname + port
	return hostname
}

func main() {
	tracer.RegisterExporterWithTracer()

	// Set up a connection to the server along with tracing stats handler
	conn, err := grpc.Dial(
		getHostName(),
		grpc.WithInsecure(),
		grpc.WithStatsHandler(new(ocgrpc.ClientHandler)),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewProductInfoClient(conn)

	// infinite loop with each loop in a 3 sec delay
	for {
		// create a context and a span for AddProduct grpc call
		addProductCtx, addProductSpan := trace.StartSpan(context.Background(), "ecommerce.client.AddProduct")

		// Contact the server and print out its response.
		name := "Sumsung S10"
		description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2019"
		price := float32(700.0)
		// send the context which we generated from trace start span
		r, err := c.AddProduct(addProductCtx, &pb.Product{Name: name, Description: description, Price: price})
		if err != nil {
			// set the status code and error message to span in case there is any error
			addProductSpan.SetStatus(
				trace.Status{Code: trace.StatusCodeInternal,
					Message: err.Error()},
			)
			log.Fatalf("Could not add product: %v", err)
		}
		// end the add product span to show the end of grpc request sent to AddProduct procedure
		addProductSpan.End()
		log.Printf("Product ID: %s added successfully", r.Value)

		// create a context and a span for GetProduct
		getProductContext, getProductSpan := trace.StartSpan(context.Background(), "ecommerce.client.GetProduct")
		product, err := c.GetProduct(getProductContext, &wrapper.StringValue{Value: r.Value})
		if err != nil {
			log.Fatalf("Could not get product: %v", err)
		}
		// end the add product span to show the end of grpc request sent to GetProduct procedure
		getProductSpan.End()
		log.Printf("Product: ", product.String())

		time.Sleep(3 * time.Second)
	}
}
