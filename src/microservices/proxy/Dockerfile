FROM golang:1.26.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o proxy-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/proxy-service .

EXPOSE 8000

CMD ["./proxy-service"]
