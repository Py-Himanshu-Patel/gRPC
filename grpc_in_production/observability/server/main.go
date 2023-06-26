// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc -I proto-gen proto-gen/product_info.proto-gen --go_out=plugins=grpc:go/product_info
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info
// Execute go run go/server/main.go

package main

import (
	"context"
	"errors"
	"fmt"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "grpc_prod/proto-gen"
	"sync"
)

const (
	port = ":50051"
)

var (
	// metrics registry. This holds all data collectors registered in the system
	reg = prometheus.NewRegistry()
	// creates standard client metrics, these are predefined metrics in lib
	grpcMetrics = grpc_prometheus.NewServerMetrics()
	//creates a custom metrics counter
	customMetricCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "product_mgt_server_handle_count",
		Help: "Total number of RPCs handled on the server",
	}, []string{"name"})
)

func init() {
	reg.MustRegister(grpcMetrics, customMetricCounter)
}

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
	customMetricCounter.WithLabelValues(in.Name).Inc()
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
	// Creates an HTTP server for Prometheus
	httpServer := &http.Server{
		Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		Addr:    fmt.Sprintf("0.0.0.0:%d", 9092),
	}
	// Creates a gRPC server with a metrics interceptor.
	// we use grpcMetrics.UnaryServerInterceptor, since we have unary service.
	// There is another interceptor called grpcMetrics.StreamServerInterceptor()
	// for streaming services.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
	)
	pb.RegisterProductInfoServer(grpcServer, &server{})
	// Initializes all standard metrics.
	grpcMetrics.InitializeMetrics(grpcServer)

	// start HTTP server for prometheus
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("Unable to start a http server")
		}
	}()
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
