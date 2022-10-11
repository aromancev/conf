FROM golang:1.19-alpine as builder

WORKDIR /app
COPY . ./
RUN go build \
  -o bin/ \
  ./cmd/gateway

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/* /app/
COPY ./cmd/gateway/config.json /config.json

CMD ["/app/gateway", "-conf", "/config.json"]
