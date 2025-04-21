# Ứng dụng Quản Lý Chi Tiêu

Ứng dụng web giúp quản lý chi tiêu cá nhân, được xây dựng bằng Go và PostgreSQL.

## Tính năng

- Đăng ký, đăng nhập tài khoản
- Thêm, xóa chi tiêu
- Tạo danh mục chi tiêu
- Tìm kiếm danh mục
- Xem thống kê chi tiêu theo ngày, tháng
- Biểu đồ chi tiêu theo ngày trong tháng
- Tải lên hình ảnh hóa đơn

## Cài đặt và Chạy

### Sử dụng Docker (Khuyến nghị)

1. Cài đặt [Docker](https://www.docker.com/products/docker-desktop/) và [Docker Compose](https://docs.docker.com/compose/install/)

2. Clone repository:
   ```
   git clone https://github.com/yourusername/QUAN-LY-CHI-TIEU.git
   cd QUAN-LY-CHI-TIEU
   ```

3. Khởi động ứng dụng:
   ```
   docker-compose up -d
   ```

4. Truy cập ứng dụng tại: http://localhost

### Chạy trực tiếp (Không sử dụng Docker)

1. Cài đặt [Go](https://golang.org/dl/) và [PostgreSQL](https://www.postgresql.org/download/)

2. Tạo database PostgreSQL:
   ```
   createdb expense_tracker
   ```

3. Cấu hình biến môi trường:
   ```
   export DB_HOST=localhost
   export DB_USER=postgres
   export DB_PASSWORD=khangttcnpm2024
   export DB_NAME=expense_tracker
   export DB_PORT=5432
   ```

4. Chạy ứng dụng:
   ```
   go run main.go
   ```

5. Truy cập ứng dụng tại: http://localhost:80

## Cấu trúc dự án

```
QUAN-LY-CHI-TIEU/
├── main.go                 # Điểm khởi đầu ứng dụng
├── pkg/
│   ├── database/           # Kết nối và migration database
│   ├── handlers/           # Xử lý HTTP request
│   ├── middleware/         # Middleware xác thực
│   └── models/             # Định nghĩa model
├── static/                 # File tĩnh (CSS, JS, hình ảnh)
│   └── uploads/            # Thư mục lưu hình ảnh upload
└── templates/              # Template HTML
```

## Môi trường phát triển

- Go 1.21+
- PostgreSQL 15
- Bootstrap 5.1
- Chart.js

## Tác giả

- Tên tác giả

## Giấy phép

Dự án này được phân phối dưới giấy phép MIT.