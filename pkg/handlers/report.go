package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// CategoryExpense lưu trữ thông tin chi tiêu theo danh mục
type CategoryExpense struct {
	CategoryID   uint
	CategoryName string
	Amount       int
	Percentage   float64
}

// GetCategoryReport trả về báo cáo chi tiêu theo danh mục
func GetCategoryReport(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy tham số thời gian từ query
	period := c.DefaultQuery("period", "month") // month, year, all

	var startDate, endDate time.Time
	now := time.Now()

	// Xác định khoảng thời gian dựa trên period
	switch period {
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	case "year":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(1, 0, 0)
	case "all":
		// Không giới hạn thời gian, lấy tất cả
		startDate = time.Time{}
		endDate = time.Date(9999, 12, 31, 23, 59, 59, 999999999, now.Location())
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Khoảng thời gian không hợp lệ"})
		return
	}

	// Truy vấn tổng chi tiêu trong khoảng thời gian
	var totalExpense int
	query := database.DB.Table("expenses").
		Where("user_id = ? AND deleted_at IS NULL", userID)

	// Thêm điều kiện thời gian nếu không phải "all"
	if period != "all" {
		query = query.Where("expense_date >= ? AND expense_date < ?", startDate, endDate)
	}

	query.Select("COALESCE(SUM(amount), 0)").Scan(&totalExpense)

	// Truy vấn chi tiêu theo danh mục
	type CategorySum struct {
		CategoryID uint
		Name       string
		Total      int
	}

	var categorySums []CategorySum

	categoryQuery := database.DB.Table("expenses").
		Select("expenses.category_id, categories.name, COALESCE(SUM(expenses.amount), 0) as total").
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ? AND expenses.deleted_at IS NULL", userID).
		Group("expenses.category_id, categories.name")

	// Thêm điều kiện thời gian nếu không phải "all"
	if period != "all" {
		categoryQuery = categoryQuery.Where("expenses.expense_date >= ? AND expenses.expense_date < ?", startDate, endDate)
	}

	categoryQuery.Scan(&categorySums)

	// Tính phần trăm và tạo kết quả
	var result []CategoryExpense
	for _, sum := range categorySums {
		percentage := 0.0
		if totalExpense > 0 {
			percentage = float64(sum.Total) / float64(totalExpense) * 100
		}

		result = append(result, CategoryExpense{
			CategoryID:   sum.CategoryID,
			CategoryName: sum.Name,
			Amount:       sum.Total,
			Percentage:   percentage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":      totalExpense,
		"categories": result,
		"period":     period,
		"startDate":  startDate,
		"endDate":    endDate,
	})
}

// GetExpensesByDateRange trả về chi tiêu trong khoảng thời gian
func GetExpensesByDateRange(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy tham số từ query
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	// Xử lý ngày bắt đầu
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày bắt đầu không hợp lệ"})
			return
		}
	} else {
		// Mặc định là đầu tháng hiện tại
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	// Xử lý ngày kết thúc
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày kết thúc không hợp lệ"})
			return
		}
		// Đặt thời gian kết thúc là cuối ngày
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	} else {
		// Mặc định là ngày hiện tại
		now := time.Now()
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	}

	// Kiểm tra ngày kết thúc phải sau ngày bắt đầu
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày kết thúc phải sau ngày bắt đầu"})
		return
	}

	// Truy vấn chi tiêu trong khoảng thời gian
	var expenses []models.Expense
	database.DB.Preload("Category").
		Where("user_id = ? AND expense_date >= ? AND expense_date <= ? AND deleted_at IS NULL",
			userID, startDate, endDate).
		Order("expense_date DESC").
		Find(&expenses)

	// Tính tổng chi tiêu
	var totalAmount int
	for _, expense := range expenses {
		totalAmount += expense.Amount
	}

	c.JSON(http.StatusOK, gin.H{
		"expenses":  expenses,
		"total":     totalAmount,
		"startDate": startDate,
		"endDate":   endDate,
	})
}

// GetIncomeExpenseComparison trả về so sánh thu nhập và chi tiêu theo tháng
func GetIncomeExpenseComparison(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy năm từ query, mặc định là năm hiện tại
	yearStr := c.DefaultQuery("year", "")
	var year int
	if yearStr != "" {
		parsedYear, err := strconv.Atoi(yearStr)
		if err != nil {
			year = time.Now().Year()
		} else {
			year = parsedYear
		}
	} else {
		year = time.Now().Year()
	}

	// Tạo mảng kết quả cho 12 tháng
	type MonthlyData struct {
		Income  int
		Expense int
		Balance int
	}
	result := make([]MonthlyData, 12)

	// Truy vấn tổng thu nhập theo tháng
	type MonthlySum struct {
		Month int
		Total int
	}

	// Lấy thu nhập theo tháng
	var incomeSums []MonthlySum
	database.DB.Table("incomes").
		Select("EXTRACT(MONTH FROM income_date) as month, COALESCE(SUM(amount), 0) as total").
		Where("EXTRACT(YEAR FROM income_date) = ? AND user_id = ? AND deleted_at IS NULL", year, userID).
		Group("EXTRACT(MONTH FROM income_date)").
		Scan(&incomeSums)

	// Lấy chi tiêu theo tháng
	var expenseSums []MonthlySum
	database.DB.Table("expenses").
		Select("EXTRACT(MONTH FROM expense_date) as month, COALESCE(SUM(amount), 0) as total").
		Where("EXTRACT(YEAR FROM expense_date) = ? AND user_id = ? AND deleted_at IS NULL", year, userID).
		Group("EXTRACT(MONTH FROM expense_date)").
		Scan(&expenseSums)

	// Khởi tạo mảng kết quả
	for i := range result {
		result[i] = MonthlyData{
			Income:  0,
			Expense: 0,
			Balance: 0,
		}
	}

	// Điền dữ liệu thu nhập
	for _, sum := range incomeSums {
		if sum.Month >= 1 && sum.Month <= 12 {
			result[sum.Month-1].Income = sum.Total
			result[sum.Month-1].Balance = result[sum.Month-1].Income - result[sum.Month-1].Expense
		}
	}

	// Điền dữ liệu chi tiêu
	for _, sum := range expenseSums {
		if sum.Month >= 1 && sum.Month <= 12 {
			result[sum.Month-1].Expense = sum.Total
			result[sum.Month-1].Balance = result[sum.Month-1].Income - result[sum.Month-1].Expense
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"year": year,
	})
}

// ShowReportPage hiển thị trang báo cáo
func ShowReportPage(c *gin.Context) {
	c.HTML(http.StatusOK, "report.html", nil)
}
