package main

import (
	pb "OrderManagement/ecommerce"
	"context"
	"fmt"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	port           = ":8000"
	orderBatchSize = 3
)

var orderMap = make(map[string]pb.Order)

type server struct {
	pb.UnimplementedOrderManagementServer
}

func (s *server) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrappers.StringValue, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	log.Println("----  AddOrder Incomming Metadata : ", md, " ------")

	if orderReq.Id == "-1" {
		log.Printf("Order ID is invalid! -> Received Order ID %s", orderReq.Id)

		errorStatus := status.New(codes.InvalidArgument, "Invalid information received")

		// Include any error details with an error type BadRequest_FieldViolation from
		// google.golang.org/genproto/googleapis/rpc/errdetails.
		ds, err := errorStatus.WithDetails(
			&epb.BadRequest_FieldViolation{
				Field:       "ID",
				Description: fmt.Sprintf("Order ID received is not valid %s : %s", orderReq.Id, orderReq.Description),
			},
		)
		// If there is some error generating the more details error response then return
		// the error response generated with status.New
		if err != nil {
			return nil, errorStatus.Err()
		}
		// return the error with details
		return nil, ds.Err()

	} else {
		orderMap[orderReq.Id] = *orderReq
		log.Println("Order : ", orderReq.Id, " -> Added")
		return &wrappers.StringValue{Value: "Order Added: " + orderReq.Id}, nil
	}
}

// unary RPC
func (s *server) GetOrder(ctx context.Context, orderId *wrappers.StringValue) (*pb.Order, error) {
	ord, found := orderMap[orderId.Value]
	if !found {
		return nil, status.Errorf(codes.NotFound, "Product does not exists : %s", orderId.Value)
	}
	return &ord, nil
}

// server streaming
func (s *server) SearchOrders(searchQuery *wrappers.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {
	for _, order := range orderMap {
		for _, itemStr := range order.Items {
			if strings.Contains(itemStr, searchQuery.Value) {
				err := stream.Send(&order)
				if err != nil {
					return fmt.Errorf("error sending message to stream : %v", err)
				}
				break
			}
		}
	}
	return nil
}

func (s *server) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	// OPTIONAL: receive and check the metadata in the stream requests
	md, _ := metadata.FromIncomingContext(stream.Context())
	log.Println("----  UpdateOrders Incomming Metadata : ", md, " ------")

	// OPTIONAL: pass the metadata back to client
	// create and send header
	header := metadata.Pairs("header-key", "header-val")
	stream.SendHeader(header)
	// create and set trailer
	trailer := metadata.Pairs("trailer-key", "trailer-val")
	stream.SetTrailer(trailer)

	ordersStr := "Updated Order IDs : "
	for {
		// Read message from the client stream.
		order, err := stream.Recv()
		// Check for end of stream.
		if err == io.EOF {
			// Finished reading the order stream.
			return stream.SendAndClose(
				&wrappers.StringValue{Value: "Orders processed " + ordersStr})
		}
		// Update order
		orderMap[order.Id] = *order

		log.Printf("Order ID ", order.Id, ": Updated")
		ordersStr += order.Id + ", "
	}
}

func (s *server) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	// Business Logic Here
	for {
		batchMarker := 1
		combinedShipmentMap := make(map[string]pb.CombinedShipment)
		for {
			orderId, err := stream.Recv()
			log.Printf("Reading Proc order : %s", orderId)
			if err == io.EOF {
				// Client has sent all the messages Send remaining shipments
				log.Printf("EOF : %s", orderId)
				for _, shipment := range combinedShipmentMap {
					if err := stream.Send(&shipment); err != nil {
						return err
					}
				}
				return nil
			}
			// error while reading client's message
			if err != nil {
				log.Println(err)
				return err
			}
			// get the destination of incoming order from it's orderId
			destination := orderMap[orderId.GetValue()].Destination
			// get the shipment
			shipment, found := combinedShipmentMap[destination]

			if found {
				ord := orderMap[orderId.GetValue()]
				shipment.OrdersList = append(shipment.OrdersList, &ord)
				combinedShipmentMap[destination] = shipment
			} else {
				comShip := pb.CombinedShipment{Id: "cmb - " + (orderMap[orderId.GetValue()].Destination), Status: "Processed!"}
				ord := orderMap[orderId.GetValue()]
				comShip.OrdersList = append(shipment.OrdersList, &ord)
				combinedShipmentMap[destination] = comShip
				log.Print(len(comShip.OrdersList), " ", comShip.GetId())
			}

			if batchMarker == orderBatchSize {
				for _, comb := range combinedShipmentMap {
					log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrdersList))
					if err := stream.Send(&comb); err != nil {
						return err
					}
				}
				batchMarker = 0
				combinedShipmentMap = make(map[string]pb.CombinedShipment)
			} else {
				batchMarker++
			}
		}
	}
}

// Server - Unary Interceptor
func orderUnaryServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	// Preprocessing logic
	// Gets info about the current RPC call by examining the args passed in
	log.Println("======= [Unary Server Interceptor - Pre Message] : ", info.FullMethod)

	// Invoking the handler to complete the normal execution of a unary RPC.
	m, err := handler(ctx, req)

	// Post processing logic
	log.Printf("======= [Unary Server Interceptor - Post Message]  : %s", m)
	return m, err
}

// Server - Streaming Interceptor
// wrappedStream wraps around the embedded grpc.ServerStream,
// and intercepts the RecvMsg and SendMsg method call.

// Wrapper stream of the grpc.ServerStream.
type wrappedStream struct {
	grpc.ServerStream
}

// Implementing the RecvMsg function of the wrapper to
// process messages received with stream RPC.
func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Wrapper] "+
		"Receive a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

// Implementing the SendMsg function of the wrapper to
// process messages sent with stream RPC
func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Wrapper] "+
		"Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

// Creating an instance of the new wrapper stream from old server stream
func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

func orderStreamServerInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Preprocessor phase.
	log.Println("====== [Server Stream Interceptor] ", info.FullMethod)
	// Invoking the streaming RPC with the wrapper stream.
	err := handler(srv, newWrappedStream(ss))
	if err != nil {
		log.Printf("RPC failed with error %v", err)
	}
	return err
}

func main() {
	initSampleData()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Registering the Interceptor at the server-side.
	s := grpc.NewServer(
		grpc.UnaryInterceptor(orderUnaryServerInterceptor),
		grpc.StreamInterceptor(orderStreamServerInterceptor),
	)
	pb.RegisterOrderManagementServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initSampleData() {
	orderMap["101"] = pb.Order{Id: "101", Items: []string{"Apple Mouse", "Mac Magic Keyboard"}, Destination: "Mountain View, CA", Price: 50.00}
	orderMap["102"] = pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	orderMap["103"] = pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	orderMap["104"] = pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	orderMap["105"] = pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	orderMap["106"] = pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}
