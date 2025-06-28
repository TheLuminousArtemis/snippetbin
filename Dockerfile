# Build stage
FROM golang:1.23 AS builder
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o web ./cmd/web

# Final stage
FROM scratch
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/web .

# Copy static assets
COPY --from=builder /app/ui/static ./ui/static
COPY --from=builder /app/ui/html ./ui/html

# Set the entrypoint
ENTRYPOINT ["./web"]

# Expose the port
EXPOSE 4000