# Cấu hình Email Reminder với Docker

## Cấu hình Docker Compose

Để sử dụng tính năng nhắc nhở chi tiêu qua email khi chạy ứng dụng với Docker, bạn cần cấu hình các biến môi trường trong file `docker-compose.yml`:

```yaml
services:
  app:
    # ... các cấu hình khác ...
    environment:
      # ... các biến môi trường khác ...
      
      # Email configuration
      - SMTP_SERVER=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USERNAME=your-email@gmail.com
      - SMTP_PASSWORD=your-app-password
      - FROM_EMAIL=your-email@gmail.com
```

## Cách cấu hình với Gmail

Nếu bạn sử dụng Gmail làm SMTP server, bạn cần tạo "App Password" thay vì sử dụng mật khẩu Gmail thông thường:

1. Đăng nhập vào tài khoản Google của bạn
2. Truy cập [Bảo mật tài khoản](https://myaccount.google.com/security)
3. Trong phần "Đăng nhập vào Google", chọn "Mật khẩu ứng dụng"
4. Tạo mật khẩu ứng dụng mới và sử dụng nó cho `SMTP_PASSWORD`

## Xây dựng và chạy ứng dụng

Trước khi xây dựng ứng dụng, hãy chạy script dọn dẹp để đảm bảo không có file trùng lặp:

### Windows
```
.\cleanup_before_build.bat
```

### Linux/Mac
```
sh cleanup_before_build.sh
```

Sau đó, xây dựng và chạy ứng dụng:

```
docker-compose build
docker-compose up -d
```

## Kiểm tra tính năng

Để kiểm tra tính năng nhắc nhở chi tiêu, bạn có thể truy cập API sau (yêu cầu đăng nhập):

```
GET /admin/test-reminder?attempt=1
```

Tham số `attempt` có thể là 1, 2, 3 hoặc 4, tương ứng với các lần nhắc nhở khác nhau.

## Lưu ý

- Đảm bảo rằng SMTP server của bạn cho phép kết nối từ ứng dụng
- Nếu sử dụng Gmail, bạn cần bật "Less secure app access" hoặc sử dụng App Password
- Tính năng này sẽ tự động chạy vào các thời điểm: 21:00, 22:00, 22:30, và 23:00