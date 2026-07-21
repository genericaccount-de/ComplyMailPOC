FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/proxy ./cmd/proxy

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/proxy /usr/local/bin/proxy
COPY config.yaml /etc/complymail/config.yaml
EXPOSE 2525
ENTRYPOINT ["proxy"]
CMD ["-config", "/etc/complymail/config.yaml"]
