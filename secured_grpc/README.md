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

## Enabling a One-Way Secured Connection (TLS)
In a one-way connection, only the client validates the server to ensure that it receives data from the intended server. When establishing the connection between the client and the server, the server shares its public certificate with the client, who then validates the received certificate of server with a CA (Certificate Authority).

To enable TLS, first we need to create the following certificates and keys:
- `server.key` A private RSA key to sign and authenticate the public key.
- `server.pem/server.crt` Self-signed X.509 public keys for distribution.

### Generate private RSA Key
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

### Generate CA and self-signed certificates
`CA = Certificate Authority`
- generate RSA key using OpenSSL, this private key is used to self sign the certificate of CA.
  ```bash
  $ openssl genrsa -aes256 -out ca.key 4096

  $ ls
  ca.key  README.md  server.key
  ```
  Do put a passphrase inorder to generate the private key. `privatekey` - passphrase, this will be asked while using this private key to generate the certificates.
- Now we can create the self signed **root CA certificate** (In cryptography and computer security, a root certificate is a public key certificate that identifies a root certificate authority. Root certificates are self-signed and form the basis of an X.509-based public key infrastructure.)
  ```bash
  $ openssl req -new -x509 -sha256 -days 3650 -key ca.key -out ca.crt
  ```
  - `-new` means new request
  - `-x509` means X.509 certificate structure instead of a cert request.
  - `-sha256` is the algo to generate the certificate
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
  The next step is to create a server private key and certificate. Unlike the previous section, we need get the certificate signed by our new Certificate Authority(CA).

### Generate server certificate


## Enabling a One-Way Secured Connection (Mutual TLS - mTLS)
