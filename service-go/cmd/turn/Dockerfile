FROM golang:1.17-alpine as builder

WORKDIR /app
COPY . ./

RUN go build \
  -o bin/ \
  ./cmd/turn

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/* /app/

CMD ["/app/turn"]
