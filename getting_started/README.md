# Getting Started with gRPC

<div align="center">
    <img src="grpc_client_server.png">
</div>

---

## Implement Server - GoLang
First, we need to generate the stubs for the service definition, then we implement the business logic of all the remote methods of the service, and finally, we create a server listening on a specified port and register the service to accept client requests.

1. Make sure `protoc` compiler is installed.
```bash
$ protoc --version
libprotoc 3.12.4
```
2. Install the gRPC library for GoLang
```bash
# download packages
go get google.golang.org/grpc
go get google.golang.org/protobuf/cmd/protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
# install packages
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
And
```bash
# put in .bashrc file
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

3. Create Go Stub/Client
```bash
protoc --go_out=. --go_opt=Mproto/product_info.proto=productinfo/server/ecommerce --go-grpc_out=. --go-grpc_opt=Mproto/product_info.proto=productinfo/server/ecommerce proto/product_info.proto
```
- In above command `--go_out=.` means when the compiler output the stub file in respective location it don't end up creating further nesting.
- `--go_opt=M{proto_file_location}={stub_file_location}`  
- In the end `proto/product_info.proto` is the location of all the proto files which are used in generating stub. Without this we get error `Missing input file.`
- Similarly genearate the `_grpc.pb.go` file as well.

The resulting structure will look like this
```bash
├── productinfo
│   ├── client
│   │   ├── go.mod
│   │   └── go.sum
│   └── server
│       ├── ecommerce
│       │   └── product_info.pb.go
│       ├── go.mod
│       └── go.sum
├── proto
│   └── product_info.proto
```

Implemented the server in `getting_started/productinfo/server/main.go`

```bash
├── productinfo
│   ├── client
│   │   ├── go.mod
│   │   └── go.sum
│   └── server
│       ├── ecommerce
│       │   ├── product_info_grpc.pb.go
│       │   └── product_info.pb.go
│       ├── go.mod
│       ├── go.sum
│       └── main.go
├── proto
│   └── product_info.proto
```
- Run server as `go run main.go` while in server directory.


### Server Stub

```go
package main

import (
	"context"
	"log"
	"net"
	pb "productinfo/server/ecommerce"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	port = ":50051"
)

// server is used to implement ecommerce/product_info.
type server struct {
	// this is required, as server is a type of ProductInfoServer
	pb.UnimplementedProductInfoServer
	productMap map[string]*pb.Product
}

// AddProduct implements ecommerce.AddProduct
func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out, err := uuid.NewV4()
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Error while generating Product ID : %s", err)
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	return &pb.ProductID{Value: in.Id}, status.New(codes.OK, "").Err()
}

// GetProduct implements ecommerce.GetProduct
func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	value, exists := s.productMap[in.Value]
	if exists {
		return value, status.New(codes.OK, "").Err()
	}
	return nil, status.Errorf(codes.NotFound, "Product does not exist : %s", in.Value)
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// create and start a new server
	s := grpc.NewServer()
	pb.RegisterProductInfoServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

## Implement Client - GoLang
Client and Server can be in different languages but I am going ahead with GoLang for consistency.

1. Create Go Stub/Client
```bash
protoc --go_out=. --go_opt=Mproto/product_info.proto=productinfo/client/ecommerce --go-grpc_out=. --go-grpc_opt=Mproto/product_info.proto=productinfo/client/ecommerce proto/product_info.proto
```
2. Hit the client while the server is running.
```bash
$ go run main.go 
2023/05/02 00:00:49 Product ID: c80f6c87-3bc1-4747-94dc-8881e3f2cef0 added successfully
2023/05/02 00:00:49 Product: id:"c80f6c87-3bc1-4747-94dc-8881e3f2cef0"  name:"Apple iPhone 11"  description:"Meet Apple iPhone 11. All-new dual-camera \n\tsystem with Ultra Wide and Night mode."
```

### Client Stub

```go
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
```