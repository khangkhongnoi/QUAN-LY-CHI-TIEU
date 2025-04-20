package main

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/handlers"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Đường dẫn templates
	absPath, _ := filepath.Abs("templates")
	log.Println("Template path:", absPath)

	// Đọc biến môi trường cho kết nối PostgreSQL
	dbHost := getEnv("DB_HOST", "localhost") // Mặc định: localhost
	dbUser := getEnv("DB_USER", "postgres")  // Mặc định: postgres
	dbPassword := getEnv("DB_PASSWORD", "khangttcnpm2024")
	dbName := getEnv("DB_NAME", "expense_tracker")
	dbPort := getEnv("DB_PORT", "5424") // Mặc định: 5424

	// Kết nối PostgreSQL
	database.InitDB(
		dbHost,
		dbUser,
		dbPassword,
		dbName,
		dbPort,
	)

	// Khởi tạo router
	router := gin.Default()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	// Routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.POST("/add", handlers.AddExpense)
	router.GET("/summary", handlers.GetSummary)
	router.GET("/expenses", handlers.GetExpenses)
	router.GET("/categories", handlers.GetCategories)
	router.GET("/daily-expenses", handlers.GetDailyExpenses)
	router.DELETE("/expenses/:id", handlers.DeleteExpense)

	// Khởi chạy server
	log.Println("Server running on :8402")
	router.Run(":8402")
}

// Hàm đọc biến môi trường, nếu không có thì dùng giá trị mặc định
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
