#!/bin/bash

# Go REST API Setup Script

# 1. Configure
echo "Configuring Go REST API..."


# 2. Install Dependencies
echo "Installing Dependencies..."
go get -u github.com/gin-gonic/gin
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/dgrijalva/jwt-go

# 3. Build the Application
echo "Building the Go REST API..."
go build ./cmd/APIRest

# 4. Run the Application
echo "Running Go REST API..."
./APIRest

# End of script
