package main

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/handlers"
	"QUAN-LY-CHI-TIEU/pkg/middleware"
	"QUAN-LY-CHI-TIEU/pkg/models"
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

	// Cập nhật auto migrate để thêm model User
	database.DB.AutoMigrate(&models.User{})

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

	// Routes công khai
	router.GET("/login", handlers.ShowLoginPage)
	router.POST("/login", handlers.Login)
	router.GET("/register", handlers.ShowRegisterPage)
	router.POST("/register", handlers.Register)
	router.GET("/logout", handlers.Logout)

	// Routes yêu cầu xác thực
	authorized := router.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		authorized.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
		authorized.POST("/add", handlers.AddExpense)
		authorized.GET("/summary", handlers.GetSummary)
		authorized.GET("/expenses", handlers.GetExpenses)
		authorized.GET("/categories", handlers.GetCategories)
		authorized.POST("/categories", handlers.AddCategory)
		authorized.GET("/daily-expenses", handlers.GetDailyExpenses)
		authorized.GET("/monthly-expenses", handlers.GetMonthlyExpenses)
		authorized.DELETE("/expenses/:id", handlers.DeleteExpense)
	}

	// Khởi chạy server
	log.Println("Server running on :80")
	router.Run(":80")
}

// Hàm đọc biến môi trường, nếu không có thì dùng giá trị mặc định
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}