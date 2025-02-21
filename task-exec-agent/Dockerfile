FROM golang:1.23.6-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o task-exec-agent main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/task-exec-agent .

CMD ["./task-exec-agent"]
