# Stage 1: Build the application
FROM golang:1.17-bullseye as builder

RUN mkdir /build && mkdir /seabird-datadog-plugin

WORKDIR /seabird-datadog-plugin
ADD ./go.mod ./go.sum ./
RUN go mod download

ADD . ./

RUN go build -v -o /build/seabird-datadog-plugin ./cmd/seabird-datadog-plugin

# Stage 2: Copy files and configure what we need
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built seabird into the container
COPY --from=builder /build/seabird-datadog-plugin /usr/local/bin

ENTRYPOINT ["/usr/local/bin/seabird-datadog-plugin"]
