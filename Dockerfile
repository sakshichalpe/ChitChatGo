FROM golang:1.23-alpine

# Install git and other dependencies
RUN apk add --no-cache git

# Set the working directory
WORKDIR /chatapp

# Copy Go module files first to leverage Docker caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the rest of the application
COPY . .

# Build the Go application
RUN go build -o main .

# Run the application
CMD ["/chatapp/main"]
