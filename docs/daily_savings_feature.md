# Tính năng tiết kiệm tự động hàng ngày

## Tổng quan

Tính năng tiết kiệm tự động hàng ngày giúp người dùng tiết kiệm tiền dựa trên chi tiêu hàng ngày của họ. Hệ thống sẽ tự động tính toán số tiền tiết kiệm dựa trên quy tắc sau:

- Hạn mức chi tiêu hàng ngày: 100,000 VND
- Nếu tổng chi tiêu trong ngày dưới 100,000 VND, phần chênh lệch sẽ được tự động thêm vào mục tiêu tiết kiệm
- Nếu tổng chi tiêu trong ngày vượt quá 100,000 VND, không có tiền tiết kiệm nào được thêm vào

## Cách hoạt động

1. Vào cuối mỗi ngày (23:59), hệ thống sẽ tự động tính toán tổng chi tiêu trong ngày của mỗi người dùng
2. Nếu tổng chi tiêu dưới 100,000 VND, hệ thống sẽ tính toán số tiền tiết kiệm: `100,000 - tổng chi tiêu`
3. Số tiền tiết kiệm sẽ được tự động thêm vào mục tiêu tiết kiệm có tên "Tiết kiệm tự động hàng ngày"
4. Nếu mục tiêu tiết kiệm này chưa tồn tại, hệ thống sẽ tự động tạo mới

## Sử dụng tính năng

### Xem thông tin về tính năng tiết kiệm tự động

```
GET /savings/daily-info
```

API này trả về thông tin về tính năng tiết kiệm tự động, bao gồm:
- Hạn mức chi tiêu hàng ngày
- Mô tả cách tính toán
- Thời gian xử lý

### Tính toán tiết kiệm thủ công

```
GET /savings/calculate-daily
```

API này cho phép người dùng tính toán và thêm tiền tiết kiệm thủ công dựa trên chi tiêu của ngày hiện tại. Kết quả trả về bao gồm số tiền tiết kiệm được thêm vào.

### Xử lý tiết kiệm cho tất cả người dùng (chỉ dành cho admin)

```
GET /admin/process-all-savings
```

API này cho phép admin kích hoạt tính toán tiết kiệm cho tất cả người dùng trong hệ thống.

## Xem tiền tiết kiệm

Người dùng có thể xem tiền tiết kiệm tự động trong phần "Mục tiêu tiết kiệm" của ứng dụng. Mục tiêu có tên "Tiết kiệm tự động hàng ngày" sẽ hiển thị tổng số tiền đã tiết kiệm được.

## Lợi ích

- Khuyến khích người dùng kiểm soát chi tiêu hàng ngày
- Tự động tích lũy tiền tiết kiệm mà không cần nỗ lực thêm
- Tạo thói quen tiết kiệm dựa trên chi tiêu thực tế
- Giúp người dùng đạt được mục tiêu tài chính dài hạn

## Lưu ý

- Tính năng này chỉ tính toán dựa trên chi tiêu được ghi lại trong hệ thống
- Để tận dụng tối đa tính năng này, người dùng nên ghi lại tất cả chi tiêu hàng ngày
- Hạn mức chi tiêu hàng ngày (100,000 VND) là cố định và không thể thay đổi