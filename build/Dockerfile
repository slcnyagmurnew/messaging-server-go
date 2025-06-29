FROM golang:1.24.3-bullseye AS builder

RUN apt-get update && apt-get install -y git build-essential make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/main ./cmd/main.go


FROM debian:bullseye-slim

# install certs for webhook.site
RUN apt-get update \
     && apt-get install -y --no-install-recommends ca-certificates \
     && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# copy the binary from builder
COPY --from=builder /app/main /app/main

EXPOSE 8080

CMD ["/app/main"]