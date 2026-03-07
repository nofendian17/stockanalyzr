FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/user-service .

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /bin/user-service /usr/local/bin/user-service

USER app
EXPOSE 50051

ENTRYPOINT ["/usr/local/bin/user-service", "serve"]
