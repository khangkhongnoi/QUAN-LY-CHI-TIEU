# Sử dụng image golang chính thức để build
FROM golang:1.24-alpine AS builder

# Cài đặt các công cụ cần thiết
RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/myapp

# Stage 2: Tạo image tối ưu
FROM alpine:latest

# Cài đặt các gói cần thiết
RUN apk --no-cache add ca-certificates tzdata

# Thiết lập múi giờ
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app
COPY --from=builder /app/myapp .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Tạo thư mục uploads
RUN mkdir -p /app/static/uploads

# Expose port
EXPOSE 80

# Biến môi trường cho database
ENV DB_HOST=postgres
ENV DB_USER=postgres
ENV DB_PASSWORD=khangttcnpm2024
ENV DB_NAME=expense_tracker
ENV DB_PORT=5432

CMD ["./myapp"]