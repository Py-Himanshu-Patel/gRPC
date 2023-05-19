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
