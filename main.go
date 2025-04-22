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

	// Cập nhật auto migrate để thêm các model
	database.DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Expense{},
		&models.IncomeCategory{},
		&models.Income{},
		&models.Budget{},
		&models.SavingGoal{},
		&models.SavingTransaction{},
	)

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
		// Trang chủ và chi tiêu
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
		authorized.GET("/expenses/:id", handlers.GetExpenseDetail)
		authorized.PUT("/expenses/:id", handlers.UpdateExpense)
		
		// Quản lý thu nhập
		authorized.GET("/income", handlers.ShowIncomePage)
		authorized.POST("/income/add", handlers.AddIncome)
		authorized.GET("/income/categories", handlers.GetIncomeCategories)
		authorized.POST("/income/categories", handlers.AddIncomeCategory)
		authorized.GET("/income/list", handlers.GetIncomes)
		authorized.DELETE("/income/:id", handlers.DeleteIncome)
		authorized.GET("/income/summary", handlers.GetIncomeSummary)
		authorized.GET("/income/monthly", handlers.GetMonthlyIncomes)
		
		// Kế hoạch ngân sách
		authorized.GET("/budget", handlers.ShowBudgetPage)
		authorized.POST("/budget/add", handlers.AddBudget)
		authorized.GET("/budget/list", handlers.GetBudgets)
		authorized.GET("/budget/summary", handlers.GetBudgetSummary)
		authorized.DELETE("/budget/:id", handlers.DeleteBudget)
		
		// Mục tiêu tiết kiệm
		authorized.GET("/saving", handlers.ShowSavingPage)
		authorized.POST("/saving/add", handlers.AddSavingGoal)
		authorized.GET("/saving/list", handlers.GetSavingGoals)
		authorized.POST("/saving/transaction", handlers.AddSavingTransaction)
		authorized.GET("/saving/:goal_id/transactions", handlers.GetSavingTransactions)
		authorized.DELETE("/saving/:id", handlers.DeleteSavingGoal)
		
		// Báo cáo và phân tích
		authorized.GET("/report", handlers.ShowReportPage)
		authorized.GET("/report/category", handlers.GetCategoryReport)
		authorized.GET("/report/date-range", handlers.GetExpensesByDateRange)
		authorized.GET("/report/comparison", handlers.GetIncomeExpenseComparison)
		
		// Tổng quan tài chính
		authorized.GET("/financial-overview", handlers.GetFinancialOverview)
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