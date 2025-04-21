package database

import (
	"log"
)

// RunMigrations thực hiện các migration cần thiết
func RunMigrations() {
	// Kiểm tra và thêm cột user_id vào bảng expenses nếu chưa tồn tại
	if !columnExists("expenses", "user_id") {
		log.Println("Thêm cột user_id vào bảng expenses...")
		
		// Thêm cột user_id với giá trị mặc định là 1 (cho dữ liệu hiện có)
		err := DB.Exec("ALTER TABLE expenses ADD COLUMN user_id bigint DEFAULT 1 NOT NULL").Error
		if err != nil {
			log.Printf("Lỗi khi thêm cột user_id: %v\n", err)
		} else {
			log.Println("Đã thêm cột user_id thành công")
		}
	}
}

// columnExists kiểm tra xem một cột có tồn tại trong bảng hay không
func columnExists(tableName, columnName string) bool {
	var count int64
	DB.Raw(`
		SELECT COUNT(1) 
		FROM information_schema.columns 
		WHERE table_name = ? 
		AND column_name = ?
	`, tableName, columnName).Scan(&count)
	
	return count > 0
}