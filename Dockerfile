# Stage 1: Build the application
FROM golang:1.22-bookworm as builder

RUN mkdir /build

WORKDIR /app

ADD ./go.mod ./go.sum ./
RUN go mod download

ADD . ./
RUN go build -v -o /build/ ./cmd/*

# Stage 2: Copy files and configure what we need
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /usr/local/bin/seabird-entrypoint.sh
COPY --from=builder /build /bin

CMD ["/usr/local/bin/seabird-entrypoint.sh"]
