# Intro to gRPC

<div align="center">
    <img src="grpc_intro.png">
</div>

- The language that we specify in the service definition is known
as an **Interface Definition Language (IDL)**.

- Using that service definition, you can generate the server-side code known as a **server skeleton**, and client-side code, known as a **client stub**.

- The methods that you specify in the service interface definition can be remotely invoked by the client side as easily as making a local function invocation.

- The network communication between the service and consumer takes place over HTTP/2.

- gRPC uses **protocol buffers** as the IDL to define the service interface. Protocol buffers are a language-agnostic, platform-neutral, extensible mechanism to serializing structured data.

## Service Definition

```protobuf
// ProductInfo.proto
syntax = "proto3";
package ecommerce;

service ProductInfo {
  rpc addProduct(Product) returns (ProductID);
  rpc getProduct(ProductID) returns (Product);
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
}

message ProductID {
  string value = 1;
}
```

- The service definition begins with specifying the protocol buffer version (proto3) that we use.
- Package names are used to prevent name clashes between protocol message types and also will be used to generate code.
- `ProductInfo` Defining the service interface of a gRPC service. 
- `addProduct(Product)` Remote method to add a product that returns the product ID as the response.
- `getProduct(ProductID)` Remote method to get a product based on the product ID.
- `Product` Definition of the message format/type of Product.
Field (name-value pair) that holds the product ID with unique field numbers that are used to identify your fields in the message binary format.
- `ProductID` User-defined type for product identification number.

# gRPC Server
- Use service definition to generate the server- or client-side code using the protocol buffer compiler **protoc**.

