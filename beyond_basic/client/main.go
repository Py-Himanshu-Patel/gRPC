package main

import (
	pb "OrderManagement/ecommerce"
	"context"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

const (
	address = "localhost:8000"
)

func orderUnaryClientInterceptor(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Preprocessor phase:  has access to the RPC request prior to sending it out to the server.
	log.Println("Method : " + method)
	// Invoking the remote method via UnaryInvoker
	err := invoker(ctx, method, req, reply, cc, opts...)
	// Postprocessor phase
	log.Println(reply)
	// return error back to grpc client
	return err
}

// a wrapper on client stream
type wrappedStream struct {
	grpc.ClientStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Client Stream Interceptor] "+
		"Receive a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Client Stream Interceptor] "+
		"Send a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

// get the wrapped stream of grpc by passing the actual stream
func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}

func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Println("======= [Client Interceptor] ", method)
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil
}

func main() {
	// Set up a connection with the server from the
	// provided address ("localhost: 8000")

	// Setting up a connection to the server.
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(clientStreamInterceptor))
	// conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

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

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()

	// Create a context to pass with remote call. This time we are using deadline
	// of 2 sec for entire call. This is different than timeout
	clientDeadline := time.Now().Add(time.Duration(2 * time.Second))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	defer cancel()

	log.Print("\n-----------------------------------------------------------------------------\n")

	// make some metadata
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"hello", "world", // key - value
	)
	// make a new context with the metadata
	newMdCtx := metadata.NewOutgoingContext(context.Background(), md)
	// append more to outgoing context
	// ctxA := metadata.AppendToOutgoingContext(newMdCtx, "k1", "v1")

	log.Print("\n-----------------------------------------------------------------------------\n")

	// Add Order
	// This is an invalid order
	order1 := pb.Order{Id: "-1", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
	// we are compressing the gRPC message here using gzip
	res, addOrderError := ordMgmtClient.AddOrder(newMdCtx, &order1, grpc.UseCompressor(gzip.Name))

	if addOrderError != nil {
		// extract the error code out of status received
		errorCode := status.Code(addOrderError)
		// match the error code for InvlidArgumet
		if errorCode == codes.InvalidArgument {
			log.Printf("Invalid Argument Error : %s", errorCode)
			// convert the error response to get more details or print as is.
			errorStatus := status.Convert(addOrderError)
			for _, d := range errorStatus.Details() {
				switch info := d.(type) {
				case *epb.BadRequest_FieldViolation:
					log.Printf("Request Field Invalid: %s", info)
				default:
					log.Printf("Unexpected error type: %s", info)
				}
			}
		} else {
			log.Printf("Unhandled error : %s ", errorCode)
		}
	} else {
		log.Print("AddOrder Response -> ", res.Value)
	}

	log.Print("\n-----------------------------------------------------------------------------\n")

	// call GetOrder method with product details, also pass the new Context
	retrievedOrder, err := ordMgmtClient.GetOrder(ctx, &wrappers.StringValue{Value: "106"})
	if err != nil {
		// If the invocation exceeds the specified deadline, it should return
		// an error of the type DEADLINE_EXCEEDED
		log.Fatalf("Could not get product: %v", err)
	}
	log.Print("GetOrder Response -> : ", retrievedOrder)

	log.Print("\n-----------------------------------------------------------------------------\n")

	// ======== server streaming client ========
	searchStream, _ := ordMgmtClient.SearchOrders(ctx, &wrappers.StringValue{Value: "Mouse"})
	for {
		searchOrder, err := searchStream.Recv()
		if err == io.EOF {
			break
		}
		// handle other possible errors
		log.Print("Search Result : ", searchOrder)
	}

	log.Print("\n-----------------------------------------------------------------------------\n")

	// ======== client streaming client ========
	// Invoking UpdateOrders remote method.
	updateStream, err := ordMgmtClient.UpdateOrders(newMdCtx)

	// retrieve header
	header, _ := updateStream.Header()
	// retrieve trailer
	trailer := updateStream.Trailer()
	log.Print("------ UpdateOrders Metadata Header : ", header, " ------")
	log.Print("------ UpdateOrders Metadata Trailer : ", trailer, " ------")

	// Handling errors related to UpdateOrders.
	if err != nil {
		log.Fatalf("%v.UpdateOrders(_) = _, %v", ordMgmtClient, err)
	}

	// Sending order update via client stream.

	// Update Orders : Client streaming scenario
	updOrder1 := pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Google Pixel Book"}, Destination: "Mountain View, CA", Price: 1100.00}
	updOrder2 := pb.Order{Id: "103", Items: []string{"Apple Watch S4", "Mac Book Pro", "iPad Pro"}, Destination: "San Jose, CA", Price: 2800.00}
	updOrder3 := pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub", "iPad Mini"}, Destination: "Mountain View, CA", Price: 2200.00}

	// Updating order 1
	if err := updateStream.Send(&updOrder1); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder1, err)
	}
	// Updating order 2
	if err := updateStream.Send(&updOrder2); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder2, err)
	}
	// Updating order 3
	if err := updateStream.Send(&updOrder3); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder3, err)
	}

	// Closing the stream and receiving the response.
	updateRes, err := updateStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v",
			updateStream, err, nil)
	}
	log.Printf("Update Orders Res : %s", updateRes)

	log.Print("\n-----------------------------------------------------------------------------\n")
	// ======== bidirectional streaming client ========
	streamProcOrder, err := ordMgmtClient.ProcessOrders(ctx)
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", ordMgmtClient, err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "102"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", ordMgmtClient, "102", err)
	}
	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "103"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", ordMgmtClient, "103", err)
	}
	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "104"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", ordMgmtClient, "104", err)
	}

	// make a channel to let the goroutine wait before fetching new record from
	// server stream before the previous one is consumed
	channel := make(chan struct{})
	// Invoke the function using Goroutines to read the messages in parallel from the service.
	go asyncClientBidirectionalRPC(streamProcOrder, channel)
	// Mimic a delay when sending some messages to the service.
	time.Sleep(time.Millisecond * 500)

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "101"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", ordMgmtClient, "101", err)
	}
	// Mark the end of stream for the client stream (order IDs).
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}
	channel <- struct{}{}
}

func asyncClientBidirectionalRPC(streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan struct{}) {
	for {
		// Read service’s messages on the client side. until the end of stream
		combinedShipment, errProcOrder := streamProcOrder.Recv()
		if errProcOrder == io.EOF {
			break
		}
		if combinedShipment != nil {
			log.Printf("Combined shipment : %s", combinedShipment.OrdersList)
		}
	}
	<-c
}
