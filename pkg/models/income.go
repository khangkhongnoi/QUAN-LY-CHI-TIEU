package models

import (
	"gorm.io/gorm"
	"time"
)

// IncomeCategory đại diện cho các loại thu nhập
type IncomeCategory struct {
	gorm.Model
	Name    string   `gorm:"unique;not null"`
	Incomes []Income `gorm:"foreignKey:CategoryID"`
}

// Income đại diện cho các khoản thu nhập
type Income struct {
	gorm.Model
	CategoryID  uint           `gorm:"not null"`
	Category    IncomeCategory `gorm:"foreignKey:CategoryID"`
	UserID      uint           `gorm:"not null"`
	Amount      int            `gorm:"not null"`
	Note        string
	IncomeDate  time.Time `gorm:"default:CURRENT_TIMESTAMP"` // Ngày thu nhập
}