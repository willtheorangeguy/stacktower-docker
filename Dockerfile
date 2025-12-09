# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /stacktower main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /stacktower /app/stacktower

# Copy the blogpost directory
COPY blogpost /app/blogpost

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["/app/stacktower", "server"]
