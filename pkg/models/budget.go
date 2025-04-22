package models

import (
	"gorm.io/gorm"
	"time"
)

// Budget đại diện cho kế hoạch ngân sách
type Budget struct {
	gorm.Model
	UserID      uint      `gorm:"not null"`
	CategoryID  uint      `gorm:"not null"`
	Category    Category  `gorm:"foreignKey:CategoryID"`
	Amount      int       `gorm:"not null"` // Số tiền dự kiến chi tiêu
	StartDate   time.Time `gorm:"not null"` // Ngày bắt đầu kế hoạch
	EndDate     time.Time `gorm:"not null"` // Ngày kết thúc kế hoạch
	Description string    // Mô tả kế hoạch
}

// BudgetSummary lưu trữ thông tin tổng hợp về ngân sách
type BudgetSummary struct {
	BudgetID      uint
	CategoryID    uint
	CategoryName  string
	BudgetAmount  int
	SpentAmount   int
	RemainingAmount int
	PercentUsed   float64
}