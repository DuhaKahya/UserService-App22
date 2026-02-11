# Build stage
FROM golang:1.24 AS builder

# Set the working directory for building
WORKDIR /app

# Copy Go module files separately for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the full application source code
COPY . .

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -o userservice .

# Adjust permissions for OpenShift:
# - Change group to root (0)
# - Give group the same permissions as the owner
RUN chgrp -R 0 /app && \
    chmod -R g=u /app && \
    chmod +x /app/userservice

# Runtime stage
FROM gcr.io/distroless/static-debian12

# Set working directory inside runtime container
WORKDIR /app

# Copy the compiled binary + permissions from the builder
COPY --from=builder /app/userservice /app/userservice

# Expose the application port (matches APP_PORT=8081)
EXPOSE 8081

# Run the binary
CMD ["./userservice"]
