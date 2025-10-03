# Use an official Go image to build the binary
FROM golang:1.25.1 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies. Dependencies are cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a minimal image based on scratch for the final image
FROM scratch

# Set the Current Working Directory inside the container
WORKDIR /

# Copy the Pre-built binary file from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/index.html .

# Command to run the executable
CMD ["/main"]
