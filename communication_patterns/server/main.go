package main

import (
	pb "OrderManagement/ecommerce"
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	port = ":8000"
)

var orderMap = make(map[string]pb.Order)

type server struct {
	pb.UnimplementedOrderManagementServer
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

func main() {
	fillData()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func fillData() {
	// put some data in map
	orderMap["106"] = pb.Order{
		Items:       []string{"Banana", "Mango"},
		Description: "Fruit Basket",
		Price:       202,
		Destination: "Fruit Market",
	}

	orderMap["101"] = pb.Order{
		Items:       []string{"Mouse", "KeyBoard"},
		Description: "Digital Basket",
		Price:       202,
		Destination: "Digital Market",
	}
}
