FROM golang:1.25.7-alpine AS builder

WORKDIR /app

# Copy pkg module first (dependency)
COPY pkg/ /pkg/

# Copy go mod files
COPY services/customer-gw-service/go.mod services/customer-gw-service/go.sum ./
RUN go mod download

# Copy source code
COPY services/customer-gw-service/ .
COPY pkg/go.sum /pkg/go.sum
RUN go mod tidy && go build -o /bin/customer-gw-service .

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /bin/customer-gw-service /usr/local/bin/customer-gw-service

USER app
EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/customer-gw-service", "serve"]
