FROM golang:1.25-alpine AS builder

ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
      -ldflags="-s -w -X github.com/m-mdy-m/atabeh/cmd.Version=${VERSION}" \
      -o /atabeh \
      ./cmd/atabeh

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /atabeh /usr/local/bin/atabeh
ENV ATABEH_DB=/data/atabeh.db
VOLUME /data

ENTRYPOINT ["/usr/local/bin/atabeh"]