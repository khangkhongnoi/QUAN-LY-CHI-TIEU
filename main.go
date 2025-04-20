package main

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/handlers"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Đường dẫn templates
	absPath, _ := filepath.Abs("templates")
	log.Println("Template path:", absPath)

	// Đọc biến môi trường
	dbURL := getEnv("DATABASE_URL", "")
	if dbURL == "" {
		// Fallback cho local development
		dbHost := getEnv("DB_HOST", "localhost")
		dbUser := getEnv("DB_USER", "postgres")
		dbPassword := getEnv("DB_PASSWORD", "khangttcnpm2024")
		dbName := getEnv("DB_NAME", "expense_tracker")
		dbPort := getEnv("DB_PORT", "5424")
		dbURL = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
			dbHost, dbUser, dbPassword, dbName, dbPort)
	}

	// Kết nối PostgreSQL
	database.InitDB(dbURL)
	// Khởi tạo router
	router := gin.Default()

	// Cấu hình CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
	router.POST("/categories", handlers.AddCategory)
	router.GET("/daily-expenses", handlers.GetDailyExpenses)
	router.DELETE("/expenses/:id", handlers.DeleteExpense)

	// Khởi chạy server
	log.Println("Server running on :8403")
	router.Run(":8403")
}

// Hàm đọc biến môi trường, nếu không có thì dùng giá trị mặc định
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
