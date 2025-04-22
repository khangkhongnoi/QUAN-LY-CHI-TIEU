package models

import (
	"gorm.io/gorm"
	"time"
)

// SavingGoal đại diện cho mục tiêu tiết kiệm
type SavingGoal struct {
	gorm.Model
	UserID      uint      `gorm:"not null"`
	Name        string    `gorm:"not null"` // Tên mục tiêu
	TargetAmount int      `gorm:"not null"` // Số tiền mục tiêu
	CurrentAmount int     `gorm:"not null;default:0"` // Số tiền hiện tại đã tiết kiệm
	Deadline    time.Time // Thời hạn hoàn thành mục tiêu
	Description string    // Mô tả mục tiêu
	Completed   bool      `gorm:"default:false"` // Trạng thái hoàn thành
}

// SavingTransaction đại diện cho các giao dịch tiết kiệm
type SavingTransaction struct {
	gorm.Model
	UserID       uint       `gorm:"not null"`
	GoalID       uint       `gorm:"not null"`
	SavingGoal   SavingGoal `gorm:"foreignKey:GoalID"`
	Amount       int        `gorm:"not null"` // Số tiền giao dịch
	TransactionDate time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Note         string
}