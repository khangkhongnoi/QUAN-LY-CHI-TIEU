# Sử dụng image golang chính thức để build
FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /myapp

# Stage 2: Tạo image tối ưu (chỉ chứa binary)
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /myapp .

CMD ["./myapp"]