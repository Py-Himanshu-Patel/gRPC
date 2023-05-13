package main

import (
	pb "OrderManagement/ecommerce"
	"context"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"time"
)

const (
	address = "localhost:8000"
)

func main() {
	// Set up a connection with the server from the
	// provided address ("localhost: 8000")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// Pass the connection and create a stub. This stub
	// instance contains all the remote methods to invoke the server.
	ordMgmtClient := pb.NewOrderManagementClient(conn)

	// Create a Context to pass with the remote call. Here
	// the Context object contains metadata such as the identity
	// of the end user, authorization tokens, and the request’s
	// deadline and it will exist during the lifetime of the request.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// call GetOrder method with product details
	retrievedOrder, err := ordMgmtClient.GetOrder(ctx, &wrappers.StringValue{Value: "106"})
	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}
	log.Print("GetOrder Response -> : ", retrievedOrder)

	log.Print("\n-----------------------------------------------------------------------------\n")

	// server streaming client
	searchStream, _ := ordMgmtClient.SearchOrders(ctx, &wrappers.StringValue{Value: "Mouse"})
	for {
		searchOrder, err := searchStream.Recv()
		if err == io.EOF {
			break
		}
		// handle other possible errors
		log.Print("Search Result : ", searchOrder)
	}
}
