#! /bin/bash
echo "Generating key..."
# openssl genrsa -out server.key 2048
openssl ecparam -genkey -name secp384rl -out server.key

echo "Generating certificate..."
openssl req -new -c509 -sha256 -key server.key -out server.out -batch days 365
