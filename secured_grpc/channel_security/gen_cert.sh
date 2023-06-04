pwd=$(basename "$PWD")
echo "Current Dir: $pwd"

# ------------- CA Private key and Cert -------------------

# generate CA private key
# with passphrase
# openssl genrsa -aes256 -out ca.key -passout file:pass.txt 4096
# without passphrase
openssl genrsa -out ca.key -passout file:pass.txt 4096

# generate CA cert
# Enter pass phrase for ca.key: privatekey
openssl req -new -x509 -sha256 -days 3650 -key ca.key -out ca.crt -subj "/C=IN/ST=KA/L=BLR"

# uncomment below line to see the text output of generated CA cert
# openssl x509 -noout -text -in ca.crt

echo "----------- CA Cert Generated -----------"

# ------ transfer CA authority cert to client and server -------
cp ca.* client/cert/
cp ca.* server/cert/

echo "-------- CA Cert Placed in client and server cert --------"

# ------- generate client and server cert using CA cert ---------
# move to server and gen server cert
cd server/cert
sh gen_server_cert.sh
cd ../..

echo "-------- Server Cert Generated Above --------"

# move to client and gen client cert
cd client/cert
sh gen_client_cert.sh
cd ../..

echo "-------- Client Cert Generated Above --------"

