# core/Dockerfile

# Use Go base image
FROM golang:1.24.3

# Set working directory
WORKDIR /src

# Copy go.mod and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
RUN go build -o aegis-api main.go

# Run the binary
CMD ["./aegis-api"]

EXPOSE 8080