# remove all certificates, keys except CA cert and key
rm -f server.key
rm -f server.crt
rm -f *.csr
rm -f *.pem
rm -f *.txt
rm -f *.cnf

# --------------- Extension File ---------------------
# put subject alternative name
echo "subjectAltName=DNS:localhost,IP:127.0.0.1" > server-ext.cnf

# ------------- Server Private key and Cert -------------------
# private key of server - unencrypted
openssl genrsa -out server.key 2048
# Cert request from server
openssl req -new -sha256 -key server.key -out server.csr -subj "/C=IN/ST=KA/L=BLR/CN=*.server.com"
# sign server request with CA pass key
openssl x509 -req -days 3650 -sha256 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 1 -out server.crt -extfile server-ext.cnf

# -------------- Verify Server Cert from CA -------------------
openssl verify -CAfile ca.crt server.crt
