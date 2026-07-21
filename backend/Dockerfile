FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/api /usr/local/bin/api
COPY config.yaml /etc/complymail/config.yaml
EXPOSE 8080
ENTRYPOINT ["api"]
CMD ["-config", "/etc/complymail/config.yaml"]
