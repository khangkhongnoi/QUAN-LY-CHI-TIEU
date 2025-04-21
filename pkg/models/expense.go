package models

import (
	"gorm.io/gorm"
	"time"
)

type Category struct {
	gorm.Model
	Name     string `gorm:"unique;not null"`
	Expenses []Expense
}

type Expense struct {
	gorm.Model
	CategoryID  uint     `gorm:"not null"`
	Category    Category `gorm:"foreignKey:CategoryID"`
	UserID      uint     `gorm:"not null"`
	Amount      int      `gorm:"not null"`
	Note        string
	ImagePath   string    // Đường dẫn đến file hình ảnh
	ImageData   string    `gorm:"type:text"` // Lưu dữ liệu base64 của hình ảnh
	ExpenseDate time.Time `gorm:"default:CURRENT_TIMESTAMP"` // Ngày chi tiêu
}
