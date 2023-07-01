// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc -I proto-gen proto-gen/product_info.proto-gen --go_out=plugins=grpc:go/product_info
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info
// Execute go run go/server/main.go

package main

import (
	"context"
	"errors"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	pb "grpc_prod/proto-gen"
	"grpc_prod/tracer"
	"log"
	"net"
	"sync"
)

const (
	port = ":50051"
)

// server is used to implement ecommerce/product_info.
type server struct {
	sync.RWMutex
	productMap map[string]*pb.Product
}

// AddProduct implements ecommerce.AddProduct
func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*wrapper.StringValue, error) {
	out, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	in.Id = out.String()
	s.Lock()
	defer s.Unlock()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	// inc the counter for given product name
	log.Printf("New product added - ID : %s, Name : %s", in.Id, in.Name)
	return &wrapper.StringValue{Value: in.Id}, nil
}

// GetProduct implements ecommerce.GetProduct
func (s *server) GetProduct(ctx context.Context, in *wrapper.StringValue) (*pb.Product, error) {
	s.Lock()
	defer s.Unlock()
	value, exists := s.productMap[in.Value]
	if exists {
		log.Printf("New product retrieved - ID : %s", in)
		return value, nil
	}

	return nil, errors.New("Product does not exist for the ID" + in.Value)
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Create a gRPC Server with gRPC interceptor.
	grpcServer := NewServer()

	pb.RegisterProductInfoServer(grpcServer, &server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func NewServer() *grpc.Server {
	// initialize jaegertracer
	jaegertracer, closer, err := tracer.NewTracer("product_mgt")
	if err != nil {
		log.Fatalln(err)
		return grpc.NewServer()
	}
	defer closer.Close()

	opentracing.SetGlobalTracer(jaegertracer)

	// initialize grpc server with chained interceptors
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			// add opentracing unary interceptor to chain
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(jaegertracer)),
		),
	)
	return server
}
