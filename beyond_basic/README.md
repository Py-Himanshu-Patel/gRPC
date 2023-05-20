# gRPC: Beyond the Basics

## Interceptors
As you build gRPC applications, you may want to execute some common logic before or after the execution of the remote function, for either client or server applications.

In gRPC you can intercept that RPC’s execution to meet certain requirements such as logging, authentication, metrics, etc., using an extension mechanism called an interceptor.

gRPC interceptors can be categorized into two types based on the type of RPC calls they intercept. 
- unary RPC you can use unary interceptors
- streaming RPC you can use streaming interceptors

Both unary and streaming interceptor can be used on client or server side.

### Server-Side Interceptors
When a client invokes a remote method of a gRPC service, you can execute a common logic prior to the execution of the remote methods by using a server-side interceptor. You can plug one or more interceptors into any gRPC server that you develop.

<div align="center">
  <img src="images/server-side-interceptor.png">
</div>

---

On the server side, the unary interceptor allows you to intercept the unary RPC call while the streaming interceptor intercepts the streaming RPC.

#### Server side - Unary interceptor

```go
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

func main() {
  ...
	// Registering the Interceptor at the server-side.
	s := grpc.NewServer(grpc.UnaryInterceptor(orderUnaryServerInterceptor))
  ...
}
```

#### Server side - Stream interceptor
The server-side streaming interceptor intercepts any streaming RPC calls that the gRPC server deals with. 

```go
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
  ...
  s := grpc.NewServer(
		grpc.UnaryInterceptor(orderUnaryServerInterceptor),
		grpc.StreamInterceptor(orderStreamServerInterceptor),
	)
  ...
}
```

### Client-Side Interceptors
When a client invokes an RPC call to invoke a remote method of a gRPC service, you can intercept those RPC calls on the client side. Applicable to both unary and streaming calls.

This is particularly useful when you need to implement certain reusable features, such as securely calling a gRPC service outside the client application code.

<div align="center">
  <img src="images/client-side-interceptor.png">
</div>

---

#### Client-Side - Unary interceptor
A client-side unary RPC interceptor is used for intercepting the unary RPC client side. `UnaryClientInterceptor` is the type for a client-side unary interceptor that has a function signature as follows.

```go
func orderUnaryClientInterceptor(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Preprocessor phase
	log.Println("Method : " + method)
	// Invoking the remote method
	err := invoker(ctx, method, req, reply, cc, opts...)
	// Postprocessor phase
	log.Println(reply)
	return err
}

func main() {
  ...
  // Setting up a connection to the server.
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor))
  ...
}
```

#### Client-Side - Stream interceptor
The client-side streaming interceptor intercepts any streaming RPC calls that the gRPC client deals with. The implementation of the client-side stream interceptor is
quite similar to that of the server side.

```go
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
  ...
	// Setting up a connection to the server.
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(clientStreamInterceptor))
  ...
}
```

## Deadlines
`Timeouts` allow you to specify how long a client application can wait for an RPC to complete before it terminates with an error. A timeout is usually specified as a duration and locally applied at each client side.

A single request may consist of multiple downstream RPCs that chain together multiple services. So we can apply timeouts, relative to each RPC, at each service invocation. Therefore, timeouts cannot be directly applied for the entire life cycle of the request. That’s where we need to use deadlines.

A `deadline` is expressed in absolute time from the beginning of a request and applied across multiple service invocations. The application that initiates the request sets the deadline and the entire request chain needs to respond by the deadline. gRPC APIs supports using deadlines.

For your RPC. For many reasons, it is always good practice to use deadlines in your gRPC applications. gRPC communication happens over the network, so there can be delays between the RPC calls and responses. 

Also, in certain cases the gRPC service itself can take more time to respond depending on the service’s business logic. When client applications are developed without using deadlines, they infinitely wait for a response for RPC requests that are initiated and resources will be held for all in-flight requests. This puts the service as well as the client at risk of running out of resources, increasing the latency of the service; this could even crash the entire gRPC service.

When client applications are developed without using deadlines, they infinitely wait for a response for RPC requests that are initiated and resources will be held for all in-flight requests. This puts the service as well as the client at risk of running out of resources, increasing the latency of the service; this could even crash the entire gRPC service.

Once the RPC call is made, the client application waits for the duration specified by the deadline; if the response for the RPC call is not received within that time, the RPC call is terminated with a `DEADLINE_EXCEEDED` error.

```go
func main() {
  ...
	clientDeadline := time.Now().Add(time.Duration(2 * time.Second))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	defer cancel()
  ...
}
```

When it comes to deadlines in gRPC, both the client and server can make their own independent and local determination about whether the RPC was successful; this means their conclusions may not match.

For instance, in our example, when the client meets the `DEADLINE_EXCEEDED` condition, the service may still try to respond. So, the service application needs to determine whether the current RPC is still valid or not. From the server side, you can also detect when the client has reached the deadline specified when invoking the RPC. Inside the AddOrder operation, you can check for `ctx.Err() == context.DeadlineExceeded` to find out whether the client has already met the deadline exceeded state, and then abandon the RPC at the server side and return an error (this is often implemented using a nonblocking select construct in Go).

## Cancellation
When either the client or server application wants to terminate the RPC this can be done by canceling the RPC. Once the RPC is canceled, no further RPC-related messaging can be done and the fact that one party has canceled the RPC is propagated to the other side.

In client RPC
```go
func main() {
  ...
	clientDeadline := time.Now().Add(time.Duration(2 * time.Second))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	
  // Canceling the RPC
  cancel()
  ...
}
```
Once you have the reference to `cancel`, you can call it at any location where you intend to terminate the RPC.

When one party cancels the RPC, the other party can determine it by checking the `context`. In this example, the server application can check whether the current context is canceled by using `stream.Context().Err() == context.Canceled`.

## Error Handling
When an error occurs, gRPC returns one of its error-status codes with an optional error message that provides more details of the error condition.

The status object is composed of an integer code and a string message that are common to all gRPC implementations for different languages.

| Code                | Number | Description                                                                                             |
|---------------------|--------|---------------------------------------------------------------------------------------------------------|
| OK                  | 0      | Success status                                                                                          |
| CANCELLED           | 1      | The operation was canceled (by the caller)                                                              |
| UNKNOWN             | 2      | Unknown error                                                                                           |
| INVALID_ARGUMENT    | 3      | The client specified an invalid argument.                                                               |
| DEADLINE_EXCEEDED   | 4      | The deadline expired before the operation could complete.                                               |
| NOT_FOUND           | 5      | Some requested entity was not found.                                                                    |
| ALREADY_EXISTS      | 6      | The entity that a client attempted to create already exists.                                            |
| PERMISSION_DENIED   | 7      | The caller does not have permission to execute the specified operation.                                 |
| RESOURCE_EXHAUSTED  | 8      | Some resource has been exhausted.                                                                       |
| FAILED_PRECONDITION | 9      | The operation was rejected because the system is not in a state required for the operation’s execution. |
| ABORTED             | 10     | The operation was aborted.                                                                              |
| OUT_OF_RANGE        | 11     | The operation was attempted past the valid range.                                                       |
| UNIMPLEMENTED       | 12     | The operation is not implemented or is not supported/enabled in this service.                           |
| INTERNAL            | 13     | Internal errors.                                                                                        |
| UNAVAILABLE         | 14     | The service is currently unavailable.                                                                   |
| DATA_LOSS           | 15     | Unrecoverable data loss or corruption.                                                                  |
| UNAUTHENTICATED     | 16     | The request does not have valid authentication credentials for the operation.                           |

```go
// server side code

func (s *server) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrappers.StringValue, error) {
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
```

```go
// client code
	// Add Order
	// This is an invalid order
	order1 := pb.Order{Id: "-1", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
	res, addOrderError := ordMgmtClient.AddOrder(ctx, &order1)

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
```

```bash
2023/05/20 11:31:16 Invalid Argument Error : InvalidArgument
2023/05/20 11:31:16 Request Field Invalid: field:"ID" description:"Order ID received is not valid -1 : "
```

## Multiplexing
