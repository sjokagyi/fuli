#!/bin/bash

# Create a playground directory
rm -rf playground
mkdir -p playground/src/services/auth
mkdir -p playground/src/services/payment
mkdir -p playground/build
mkdir -p playground/node_modules

# Create some source files
echo "package auth" > playground/src/services/auth/auth.go
echo "package payment" > playground/src/services/payment/payment.go
echo "# Fuli Config" > playground/config.yaml

# Create a binary file (simulate with random data)
head -c 100 /dev/urandom > playground/app.exe

# Create an ignore file
cat <<EOF > playground/.contextignore
# Ignore build artifacts
/build
*.exe

# Ignore dependencies
node_modules/

# Ignore secrets
.env
EOF

# Create a secret file that should be ignored
echo "SECRET_KEY=12345" > playground/.env

# Create a file in build that should be ignored
echo "binary data" > playground/build/main.o

echo "Playground created in ./playground"