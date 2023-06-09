# Secure gRPC

Before starting know what is SSL encryption

- The client contacts the server using a secure URL (HTTPS…).
- The server sends the client its certificate and public key.
- The client verifies this with a Trusted Root Certification Authority to ensure the certificate is legitimate.
- The client and server negotiate the strongest type of encryption that each can support.
- The client encrypts a session (secret) key with the server’s public key, and sends it back to the server.
- The server decrypts the client communication with its private key, and the session is established.
- The session key (symmetric encryption) is now used to encrypt and decrypt data transmitted between the client and server.
- Which all CA to trust are installed in each machine already, in case we create our own CA then it's our work to get that CA certificate to each of the client machine manually and install them.
- To make sure the public key server (example.com) is sending is actual, the client call CA and ask whether the sent publi key of server (example.com) belong to example.com or it's a man in middle attack. 

A in depth explanation of Digital Certificate, Self Sign Certificates, CA : https://www.youtube.com/watch?v=qXLD2UHq2vk

### Note: 
Why don't the client send data to server by encrypting it with the public key of server and server can decrypt it with it's private key. Because that's slow. Thus server and client end up negotiating on most strongest algo both can support with a symmetric key (same key for both server and client). That's fast in enc and decryption.

### Script to Generate: CA, Server Cert, Client Cert

```bash
channel_security$ sh gen_cert.sh

Current Dir: channel_security
----------- CA Cert Generated -----------
-------- CA Cert Placed in client and server cert --------
Certificate request self-signature ok
subject=C = IN, ST = KA, L = BLR, CN = *.server.com
server.crt: OK
-------- Server Cert Generated --------
Certificate request self-signature ok
subject=C = IN, ST = KA, L = BLR
C = IN, ST = KA, L = BLR
error 18 at 0 depth lookup: self-signed certificate
error client.crt: verification failed
-------- Client Cert Generated --------
```

## Generate CA and self-signed certificates

`CA = Certificate Authority`
- generate RSA key using OpenSSL, this private key is used to self sign the certificate of CA.
  ```bash
  $ openssl genrsa -aes256 -out ca.key 4096

  $ ls
  ca.key  README.md  server.key
  ```
  Put a passphrase inorder to generate the private key. `privatekey` - passphrase, this will be asked while using this private key to generate the certificates. In case we like to go without passphrase key, then use this command `openssl genrsa -out ca.key 4096`.
- Now we can create the self signed **root CA certificate** (In cryptography and computer security, a root certificate is a public key certificate that identifies a root certificate authority. Root certificates are self-signed and form the basis of an X.509-based public key infrastructure.)
  ```bash
  $ openssl req -new -x509 -sha256 -days 3650 -key ca.key -out ca.crt
  ```
  - `-new` means new request
  - `-x509` means X.509 certificate structure instead of a cert request.
  - `-aes256` is the algo to generate the certificate, skip this flag in order to generate a pri. key which don't need password.
  - `-days` days of validity
  - `-key` is the key we gen before to that will be go inside certificates.
  - `-out` is the name of output file
- This is an interactive process where all fields are optional but `Common Name` or `FQDN` is necessary.
  ```bash
  Country Name (2 letter code) [AU]:IN
  State or Province Name (full name) [Some-State]:KA
  Locality Name (eg, city) []:Bengaluru
  Organization Name (eg, company) [Internet Widgits Pty Ltd]:LLP
  Organizational Unit Name (eg, section) []:Engineering
  Common Name (e.g. server FQDN or YOUR name) []:my-domain.com
  Email Address []:
  ```
- Check the generated certificate as (This will have the public key of CA).
  ```bash
  $ openssl x509 -noout -text -in ca.crt 
  Certificate:
      Data:
          Version: 3 (0x2)
          Serial Number:
              74:51:49:fe:d0:da:f7:56:80:83:66:b0:af:0e:85:8a:7d:79:0b:c8
          Signature Algorithm: sha256WithRSAEncryption
          Issuer: C = IN, ST = KA, L = Bengaluru, O = LLP, OU = Engineering, CN = my-domain.com
          Validity
              Not Before: May 24 05:57:17 2023 GMT
              Not After : May 21 05:57:17 2033 GMT
          Subject: C = IN, ST = KA, L = Bengaluru, O = LLP, OU = Engineering, CN = my-domain.com
          Subject Public Key Info:
              Public Key Algorithm: rsaEncryption
                  Public-Key: (4096 bit)
                  Modulus:
                      00:e3:b6:b3:cf:2c:59:ef:3d:88:71:d3:88:aa:f9:
                      ...
                  Exponent: 65537 (0x10001)
          X509v3 extensions:
              X509v3 Subject Key Identifier: 
                  55:C2:1F:F5:E7:2E:13:63:B3:DD:2F:3F:CF:C9:1E:12:0B:15:84:A8
              X509v3 Authority Key Identifier: 
                  55:C2:1F:F5:E7:2E:13:63:B3:DD:2F:3F:CF:C9:1E:12:0B:15:84:A8
              X509v3 Basic Constraints: critical
                  CA:TRUE
      Signature Algorithm: sha256WithRSAEncryption
      Signature Value:
      ...
  ```
  The values provided interactively can also be given via `-subj` flag thus reduce the burden of typing all data. Also the data provided here 
  is not important and can be skipped by typing `.` instead of text.
  The next step is to create a server private key and certificate. Unlike the previous section, we need get the certificate signed by our new Certificate Authority(CA).

## Generate server certificate
```bash
$ openssl genrsa -out server.key 2048
Generating RSA key with 2048 bits

$ ls
README.md  server.key
```

- `genrsa` is the algorithm for generating the RSA keys
- `server.key` is the file which holds the key
- `2048` size of the key, default is 512 which is prone to brute force.
- Here you can also add a passphrase to the key. So you need the passphrase whenever you need to use the key. In this example, we are not going to add a passphrase to the key.
- We will use this `server.key` on out gRPC server.
- We did not supply any algo to encrypt the private key of server, Thus while we print (`cat`) the key we see the actual private key, without any encryption. Also the process do not ask for any passphrase. In CA private key we will use a algo to secure it, but once we generate a private key with a passphrase then we need to put the used passphrase every time we use the private key.


To enable TLS, first we need to create the following certificates and keys:
- `server.key` A private RSA key to sign and authenticate the public key.
- `server.pem/server.crt` Self-signed X.509 public keys for distribution.

Once we have the server private key, we can proceed to create a `Certificate Signing Request (CSR)`. This is a formal request asking a CA to sign a certificate, and it contains the public key of the entity requesting the certificate and some information about the entity. This will ensure all client who connect to the server can verify the public key of server from the CA.

- create a certificate signing request
  ```bash
  $ openssl req -new -sha256 -key server.key -out server.csr

  Country Name (2 letter code) [AU]:IN
  State or Province Name (full name) [Some-State]:KA
  Locality Name (eg, city) []:BLR
  Organization Name (eg, company) [Internet Widgits Pty Ltd]:LLP
  Organizational Unit Name (eg, section) []:Engineering
  Common Name (e.g. server FQDN or YOUR name) []:*.my-server.com
  Email Address []:

  Please enter the following 'extra' attributes
  to be sent with your certificate request
  A challenge password []:privateserver
  An optional company name []:LLP

  $ ls
  ca.crt  ca.key  README.md  server.csr  server.key
  ```
- After a CSR is generated, we can sign the request and generate the certificate using our own CA certificate. Normally, the CA and the certificate requester are two different companies who don’t want to share their private keys. 
- use our root CA to sign the CSR and create server certificate.
  ```bash
  $ openssl x509 -req -days 3650 -sha256 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 1 -out server.crt
  Certificate request self-signature ok
  subject=C = IN, ST = KA, L = BLR, O = LLP, OU = Engineering, CN = *.my-server.com
  Enter pass phrase for ca.key:

  $ ls
  ca.crt  ca.key  README.md  server.crt  server.csr  server.key
  ```
- we have created server key(server.key) and server certificate(server.crt). We can use them to enable mutual TLS in server side later

## Generate client key and certificate
Generating the client certificate is very similar to creating the server certificate.
```bash
$ openssl genrsa -out client.key 2048

$ openssl req -new -key client.key -out client.csr
Country Name (2 letter code) [AU]:IN
State or Province Name (full name) [Some-State]:KA
Locality Name (eg, city) []:BLR
Organization Name (eg, company) [Internet Widgits Pty Ltd]:LLP
Organizational Unit Name (eg, section) []:.
Common Name (e.g. server FQDN or YOUR name) []:*.my-client.com
Email Address []:

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:
An optional company name []:

$ openssl x509 -req -days 3650 -sha256 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 2 -out client.crt
Certificate request self-signature ok
subject=C = IN, ST = KA, L = BLR, O = LLP, CN = *.my-client.com
Enter pass phrase for ca.key:

$ ls
ca.crt  ca.key  client.crt  client.csr  client.key  README.md  server.crt  server.csr  server.key
```
You are about to be asked to enter information that will be incorporated into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN. There are quite a few fields but you can leave some blank
For some fields there will be a default value, If you enter '.', the field will be left blank.


## Convert server/client keys to pem format
```bash
$ openssl pkcs8 -topk8 -inform pem -in server.key -outform pem -nocrypt -out server.pem
$ openssl pkcs8 -topk8 -inform pem -in client.key -outform pem -nocrypt -out client.pem

$ ls
ca.crt  client.crt  client.key  README.md   server.csr  server.pem
ca.key  client.csr  client.pem  server.crt  server.key
```

## A proto file for following examples
```protobuf
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
  float price = 4;
}

message ProductID {
  string value = 1;
}
```

## Enabling a One-Way Secured Connection (TLS)
In a one-way connection, only the client validates the server to ensure that it receives data from the intended server. When establishing the connection between the client and the server, the server shares its public certificate with the client, who then validates the received certificate of server with a CA (Certificate Authority).

```bash
protoc --go_out=. --go_opt=MproductInfo.proto=server/ecommerce --go-grpc_out=. --go-grpc_opt=MproductInfo.proto=server/ecommerce productInfo.proto

protoc --go_out=. --go_opt=MproductInfo.proto=client/ecommerce --go-grpc_out=. --go-grpc_opt=MproductInfo.proto=client/ecommerce productInfo.proto
```

### Modify Server - One-Way TLS
```go
package main

import (
	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	pb "server/ecommerce"
)

var (
	port    = ":50051"
	crtFile = "cert/server.crt"
	keyFile = "cert/server.key"
)

type server struct {
	pb.UnimplementedProductInfoServer
	productMap map[string]*pb.Product
}

func main() {
	// Read and parse a public/private key pair and create
	// a certificate to enable TLS.
	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load key pair: %s", err)
	}
	// Enable TLS for all incoming connections by adding
	// certificates as TLS server credentials.
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}
	// Create a new gRPC server instance by passing TLS server credentials.
	s := grpc.NewServer(opts...)

	// Register the implemented service to the newly created
	// gRPC server by calling generated APIs.
	pb.RegisterProductInfoServer(s, &server{})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Bind the gRPC server to the listener and start listening
	// to incoming messages on the port (50051)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

### Modify Client - One-Way TLS
In order to get the client connected, the client needs to have the server’s self-certified public key.
```go
package main

import (
	// pb "client/ecommerce"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

var (
	address  = "localhost:50051"
	hostname = "localhost"
	crtFile  = "cert/server.crt"
)

func main() {
	// Read and parse a public certificate and create a certificate to enable TLS.
	creds, err := credentials.NewClientTLSFromFile(crtFile, hostname)
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProductInfoClient(conn)
	// Skip RPC method invocation.
}
```


## Enabling a Two-Way Secured Connection (Mutual TLS - mTLS)

The main intent of an mTLS connection between client and server is to have control of clients who connect to the server. Unlike a one-way TLS connection, the server is configured to accept connections from a limited group of verified clients.

Parties share their public certificates with each other and validate the other party. The basic flow of connection is as follows:

1. Client sends a request to access protected information from the server.
2. The server sends its X.509 certificate to the client.
3. Client validates the received certificate through a CA for CA-signed certificates.
4. If the verification is successful, the client sends its certificate to the server.
5. Server also verifies the client certificate through the CA.
6. Once it is successful, the server gives permission to access protected data.

We need to create a CA with self-signed certificates, we need to create certificate-signing requests for both client and server, and we need to sign them using our CA. As in the previous one-way secured connection, we can use the OpenSSL tool to generate keys and certificates.

- `server.key` - Private RSA key of the server.
- `server.crt` - Public certificate of the server.
- `client.key` - Private RSA key of the client.
- `client.crt` - Public certificate of the client.
- `ca.crt` - Public certificate of a CA used to sign all public certificates.

### Server Code

```go
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"net"
	pb "server/ecommerce"
)

var (
	port    = ":50051"
	crtFile = "cert/server.crt"
	keyFile = "cert/server.key"
	caFile  = "cert/ca.crt"
)

type server struct {
	pb.UnimplementedProductInfoServer
	productMap map[string]*pb.Product
}

// AddProduct implements ecommerce.AddProduct
func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	return &pb.ProductID{Value: in.Id}, nil
}

// GetProduct implements ecommerce.GetProduct
func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	value, exists := s.productMap[in.Value]
	if exists {
		return value, nil
	}
	return nil, errors.New("Product does not exist for the ID" + in.Value)
}

func main() {
	// Read and parse a public/private key pair and create
	// a certificate to enable TLS.
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load key pair: %s", err)
	}

	// Create a certificate pool from the CA.
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA to the certificate pool.
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certificate")
	}

	// Enable TLS for all incoming connections by creating TLS credentials.
	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		grpc.Creds(
			credentials.NewTLS(&tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{certificate},
				ClientCAs:    certPool,
			},
			)),
	}
	// Create a new gRPC server instance by passing TLS server credentials.
	s := grpc.NewServer(opts...)

	// Register the implemented service to the newly created
	// gRPC server by calling generated APIs.
	pb.RegisterProductInfoServer(s, &server{})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Bind the gRPC server to the listener and start listening
	// to incoming messages on the port (50051)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

### Client Code
```go
package main

import (
	// pb "client/ecommerce"
	pb "client/ecommerce"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
)

var (
	address  = "localhost:50051"
	hostname = "localhost"
	crtFile  = "cert/client.crt"
	keyFile  = "cert/client.key"
	caFile   = "cert/ca.crt"
)

func main() {
	// Create X.509 key pairs directly from the client certificate and key.
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}
	// Create a certificate pool from the CA.
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}
	// Append the client certificates from the CA to the certificate pool.
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:   hostname, // NOTE: this is required!
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProductInfoClient(conn)
	fmt.Println("Connection Established : ", c)
	// Skip RPC method invocation.
}
```
```bash
$ go run main.go 
2023/06/04 12:01:22 Product ID: 6c41bcb7-02a1-11ee-88be-902e16d6a0f2 added successfully
2023/06/04 12:01:22 Product: id:"6c41bcb7-02a1-11ee-88be-902e16d6a0f2"  name:"Sumsung S10"  description:"Samsung Galaxy S10 is the latest smart phone, launched in February 2019"  price:700
```

## Authenticating gRPC Calls

### Basic Auth - Username + Password
In the OAuth 2.0 flow, there are four main characters: the client, the authorization server, the resource server, and the resource
owner. The client wants to access the resource in a resource server. To access the resource, the client needs to get a token (which is an
arbitrary string) from the authorization server. This token must be of a proper length and should not be predictable. Once the client
receives the token, the client can send a request to the resource server with the token. The resource server then talks to the corre‐
sponding authorization server and validates the token. If it is validated by this resource owner, the client can access the resource

### Token Auth (OAuth 2.0) - Bearer Token Based Auth
- Read from Book

### JWT Auth - JWT Token Based Auth
JWT is preferred over Basic or Token Auth because
- Resource server doesn't need to talk to the authentication server to validate the client token.
- JWT can carry time period till it is valid, other auth methods can't.
- Authentication need Secure Channel as requirement. So make sure TLS certificate are installed 
on both client and server before implementing JWT or any other token access method.
- If we want to skip the TLS certificate requirement then workaround is to send the JWT in request metadata 
by intercepting the request before sending it. And validate the JWT at server end by intercepting the request again. 

#### JWT - Symmetric Key
- Auth server generates the jwt-secret and use it to make signature
- signature at end of each JWT = SHA256(header+"."+payload, jwt-secret)
- Thus if someone changed the payload then he don't have the jwt-secret and 
hence can't generate the expected signature with new data.
- Each application server (validator of jwt) have this jwt-secret.
- Receiver of JWT can generate the signature using header+payload+jwt-secret 
and verfiy result against the one present in JWT Token, if both match then it's valid Token.
- Downside is each app server need the jwt-secret and if one service end up compromising then 
secret need to be updated for all

#### JWT - Asymmetric Key
- Auth server will generate the public and private key. Server will enc the header+payload with private 
key and append the signature at end of JWT Token.
- A public key can decrypt the message of corresponding private key.
- Each application sever will have access to public key, which can be used to decrypt the private key signature.
- If JWT got tempered then attacker can no longer generate the sign as they don't have the private key.
- Even if the public key of one server got compromised we don't need to worry as this is public key. Also private keyi

Refer: https://www.youtube.com/watch?v=kVpB-uH6X-s

```bash
$ go get github.com/dgrijalva/jwt-go
```


### Google Token-Based Auth
- Read from Book

## Summary
There are two types of credential supports in gRPC, channel and call. Channel credentials are attached to the channels such as TLS, etc. Call credentials are attached to the call, such as OAuth 2.0 tokens, basic authentication, etc. We even can apply both credential types to the gRPC application. For example, we can have TLS enable the connection between client and server and also attach credentials to each RPC call made on the connection.
