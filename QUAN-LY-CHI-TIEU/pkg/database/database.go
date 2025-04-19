package database

import (
	"QUAN-LY-CHI-TIEU/pkg/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(host, user, password, dbname, port string) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		host, user, password, dbname, port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	autoMigrate()
	seedCategories()
}

func autoMigrate() {
	DB.AutoMigrate(
		&models.Category{},
		&models.Expense{},
	)
}

func seedCategories() {
	categories := []string{
		"Ăn sáng",
		"Trà lai",
		"Ăn trưa",
		"Nước mía",
		"Ăn chiều",
		"Ăn tối",
		"Đồ ăn khác",
		"Nước đá",
		"Thuốc lá",
		"Bia", "Rượu", "Cà phê",
		"Tiền trọ tiền điện nước", "Internet",
		"Mua sắm quần áo", "Mua sắm đồ gia dụng",
		"Thuốc", "Dịch vụ y tế",
		"Shoppee", "Thay nhớt",
		"Thẻ tín dụng",
		"Khác",
	}

	for _, name := range categories {
		DB.FirstOrCreate(&models.Category{}, models.Category{Name: name})
	}
}
