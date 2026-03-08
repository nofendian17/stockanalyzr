FROM golang:1.25.7-alpine AS builder

WORKDIR /app

# Copy pkg module first (dependency)
COPY pkg/ /pkg/

# Copy go mod files
COPY services/user-service/go.mod services/user-service/go.sum ./
RUN go mod download

# Copy source code
COPY services/user-service/ .
COPY pkg/go.sum /pkg/go.sum
RUN go mod tidy && go build -o /bin/user-service .

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /bin/user-service /usr/local/bin/user-service
COPY --from=builder /app/migrations ./migrations

USER app
EXPOSE 50051

ENTRYPOINT ["/usr/local/bin/user-service"]
CMD ["serve"]
