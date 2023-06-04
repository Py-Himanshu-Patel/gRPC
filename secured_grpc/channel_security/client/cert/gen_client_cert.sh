# remove all certificates, keys except CA cert and key
rm -f client.key
rm -f client.crt
rm -f *.csr
rm -f *.pem
rm -f *.txt
rm -f *.cnf

# --------------- Extension File ---------------------
# put subject alternative name
echo "subjectAltName=DNS:localhost,IP:127.0.0.1" > client-ext.cnf

# ------------- Client Private key and Cert -------------------
# private key for client
openssl genrsa -out client.key 2048
# Cert request from client
openssl req -new -key client.key -out client.csr -subj "/C=IN/ST=KA/L=BLR"
# sign client request with CA pass key
openssl x509 -req -days 3650 -sha256 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 2 -out client.crt -extfile client-ext.cnf

# -------------- Verify Client Cert from  -------------------
openssl verify -CAfile ca.crt client.crt
