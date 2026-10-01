FROM golang:1.26.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o events-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/events-service .

EXPOSE 8082

CMD ["./events-service"]
