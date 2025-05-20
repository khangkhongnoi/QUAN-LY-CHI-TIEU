# Hướng dẫn sử dụng tính năng nhắc nhở chi tiêu qua email

## Tổng quan

Tính năng nhắc nhở chi tiêu qua email sẽ tự động kiểm tra xem người dùng đã thêm chi tiêu trong ngày hay chưa. Nếu chưa, hệ thống sẽ gửi email nhắc nhở vào các thời điểm cố định trong ngày:

- 21:00: Nhắc nhở lần đầu
- 22:00: Nhắc nhở lần hai
- 22:30: Nhắc nhở lần ba
- 23:00: Nhắc nhở cuối cùng

## Cấu hình

Để sử dụng tính năng này, bạn cần cấu hình các biến môi trường sau:

```
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=your-email@gmail.com
```

### Lưu ý khi sử dụng Gmail

Nếu bạn sử dụng Gmail làm SMTP server, bạn cần tạo "App Password" thay vì sử dụng mật khẩu Gmail thông thường:

1. Đăng nhập vào tài khoản Google của bạn
2. Truy cập [Bảo mật tài khoản](https://myaccount.google.com/security)
3. Trong phần "Đăng nhập vào Google", chọn "Mật khẩu ứng dụng"
4. Tạo mật khẩu ứng dụng mới và sử dụng nó cho `SMTP_PASSWORD`

## Cách hoạt động

1. Hệ thống sẽ tự động chạy kiểm tra vào các thời điểm đã cấu hình
2. Đối với mỗi người dùng, hệ thống sẽ kiểm tra xem họ đã thêm chi tiêu nào trong ngày hay chưa (dựa trên trường `expense_date`)
3. Nếu chưa có chi tiêu nào được thêm, hệ thống sẽ gửi email nhắc nhở đến địa chỉ email của người dùng
4. Email nhắc nhở sẽ có nội dung khác nhau tùy thuộc vào số lần nhắc nhở

## Kiểm tra thủ công

Để kiểm tra tính năng này thủ công, bạn có thể truy cập API sau (yêu cầu đăng nhập):

```
GET /admin/test-reminder?attempt=1
```

Tham số `attempt` có thể là 1, 2, 3 hoặc 4, tương ứng với các lần nhắc nhở khác nhau.

## Lưu ý

- Người dùng cần có địa chỉ email hợp lệ trong hệ thống để nhận được nhắc nhở
- Tính năng này chỉ kiểm tra chi tiêu được thêm trong ngày hiện tại
- Nếu người dùng đã thêm ít nhất một chi tiêu trong ngày, họ sẽ không nhận được email nhắc nhở