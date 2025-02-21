FROM golang:1.23.6-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o backend-api-server main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/backend-api-server .

EXPOSE 8080

CMD ["./backend-api-server"]
