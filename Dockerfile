# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates nodejs npm

# Install Go tools
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

WORKDIR /app

# Copy all files
COPY . .

# Install npm dependencies and build CSS
RUN npm install && npm run build-css-prod

# Generate code
RUN sqlc generate && templ generate

# Build binary
RUN CGO_ENABLED=0 go build -o server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and static files
COPY --from=builder /app/server .
COPY --from=builder /app/web/static ./web/static

EXPOSE 80

CMD ["./server"]