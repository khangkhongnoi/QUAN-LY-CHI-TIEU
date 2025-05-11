package models

import (
	"time"
	"gorm.io/gorm"
)

// SavingChallenge đại diện cho một thách thức tiết kiệm
type SavingChallenge struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TargetAmount int       `json:"target_amount"`
	CurrentAmount int      `json:"current_amount"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CategoryID  uint      `json:"category_id"`
	Status      string    `json:"status"` // "active", "completed", "failed"
}

// ExpenseForecast đại diện cho dự báo chi tiêu
type ExpenseForecast struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	CategoryID  uint      `json:"category_id"`
	Month       int       `json:"month"`
	Year        int       `json:"year"`
	Amount      int       `json:"amount"`
	Confidence  float64   `json:"confidence"` // 0-1, độ tin cậy của dự báo
}

// AIRecommendation đại diện cho đề xuất tiết kiệm từ AI
type AIRecommendation struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	CategoryID  uint      `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PotentialSaving int   `json:"potential_saving"`
	Implemented bool      `json:"implemented"` // Người dùng đã thực hiện đề xuất chưa
}

// ExpensePattern đại diện cho mẫu chi tiêu được phát hiện
type ExpensePattern struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	PatternType string    `json:"pattern_type"` // "recurring", "seasonal", "impulse"
	Description string    `json:"description"`
	CategoryIDs string    `json:"category_ids"` // Danh sách ID danh mục, phân tách bằng dấu phẩy
	AverageAmount int     `json:"average_amount"`
	Frequency   string    `json:"frequency"` // "daily", "weekly", "monthly"
}

// BudgetWarning đại diện cho cảnh báo ngân sách
type BudgetWarning struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	BudgetID    uint      `json:"budget_id"`
	CategoryID  uint      `json:"category_id"`
	WarningType string    `json:"warning_type"` // "approaching_limit", "exceeded"
	Threshold   int       `json:"threshold"` // Phần trăm ngưỡng cảnh báo (80, 90, 100)
	Message     string    `json:"message"`
	Dismissed   bool      `json:"dismissed"` // Người dùng đã bỏ qua cảnh báo
}

// ReceiptScan đại diện cho hóa đơn được quét
type ReceiptScan struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	ImagePath   string    `json:"image_path"`
	ImageData   string    `json:"image_data"` // Base64 encoded image
	Status      string    `json:"status"` // "pending", "processed", "failed"
	TotalAmount int       `json:"total_amount"`
	MerchantName string   `json:"merchant_name"`
	ReceiptDate time.Time `json:"receipt_date"`
	Items       string    `json:"items"` // JSON string of items
}