# api/Dockerfile

# Use Go base image
FROM golang:1.24.3

# Set working directory
WORKDIR /src

# Copy go.mod and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Copy certificates
COPY ./certs /src/api/certs

# Build the binary
RUN go build -o aegis-api main.go

# Add this before the CMD line
RUN find /src -name "*.pem" -o -name "certs" | sort
RUN ls -la /src/api/certs/ 2>/dev/null || echo "certs directory not found"

# Run the binary
CMD ["./aegis-api"]

EXPOSE 8443