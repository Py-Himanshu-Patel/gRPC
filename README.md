# gRPC - Up and Running

[Github Repo of Book - Code Samples](https://github.com/grpc-up-and-running/samples)

## Turn on the logging in GoLang GRPC
```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99 
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

## Quick Notes and Practice Samples
1. [Intro to gRPC](intro/README.md)
2. [Getting Started gRPC](getting_started/README.md)
    - Client Stub
    - Sever Stub
    - Protoc Compiler
3. [gRPC Communication Patterns](communication_patterns/README.md)
    - Simple RPC (Unary RPC)
    - Stream RPC
      - Server Streaming
      - Client Streaming
    - Bidirectional Stream
4. [gRPC - Under the hood](under_the_hood/README.md)
    - Bi-Directional flow of information
    - Using HTTP/2 for communication
    - Length-Prefixed Message Framing (why gRPC is fast)
5. [gRPC: Beyond the Basics](beyond_basic/README.md)
    - Interceptors
    - Server Side - Unary Interceptor
    - Server Side - Stream Interceptor
    - Client Side - Unary Interceptor
    - Client Side - Stream Interceptor
    - Deadlines
    - Cancellation
    - Error Handling
    - Multiplexing (multiple service on same gRPC server)
    - Metadata (sending and receiving)
    - Load Balancing
    - Compression of gRPC message
6. [Secured gRPC](secured_grpc/README.md)
    - Generating Certificates for Client | Server | CertificateAuthority
    - One-Way Secured Connection (TLS)
    - Two-Way Secured Connection (Mutual TLS - mTLS)
7. [gRPC in Production](grpc_in_production/README.md)
    - Testing a gRPC Server
    - Testing a gRPC Client
    - Load Testing
    - Continuous Integration
    - Deployment (Docker)
    - Deployment Kubernetes - Deployment + Services + )
    - Metrics (Prometheus | OpenCensus)
    - Logs
    - Tracing (zipkin)
