# ==========================================
# Stage 1: Build the React Frontend
# ==========================================
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web

# Install dependencies
COPY web/package.json web/package-lock.json* ./
RUN npm ci

# Copy source and build
COPY web/ .
RUN npm run build

# ==========================================
# Stage 2: Build the Go Backend
# ==========================================
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o arena main.go

# ==========================================
# Stage 3: Final Runtime Image
# ==========================================
FROM alpine:3.19
WORKDIR /app

# Install required runtime dependencies for Nmap
RUN apk add --no-cache nmap nmap-scripts ca-certificates tzdata

# Copy the compiled Go binary
COPY --from=backend-builder /app/arena .

# Copy the configuration file and vlan mappings
COPY config.yaml .
COPY vlans.json .

# Copy the built React application
COPY --from=frontend-builder /app/web/dist ./web/dist

# Ensure the binary is executable
RUN chmod +x /app/arena

# Expose the default HTTP port
EXPOSE 8080

# Run the backend in daemon mode by default
ENTRYPOINT ["/app/arena"]
CMD ["/app/config.yaml"]
