package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// AddIncome xử lý thêm khoản thu nhập mới
func AddIncome(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		CategoryID  string `form:"income_category_id" binding:"required"`
		NewCategory string `form:"new_income_category"`
		Amount      string `form:"income_amount" binding:"required"`
		Note        string `form:"income_note"`
		IncomeDate  string `form:"income_date"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Xử lý category mới
	categoryID, err := strconv.ParseUint(request.CategoryID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	// Nếu chọn tạo danh mục mới
	if categoryID == 0 && request.NewCategory != "" {
		newCategory := models.IncomeCategory{Name: request.NewCategory}
		result := database.DB.FirstOrCreate(&newCategory, newCategory)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		categoryID = uint64(uint(newCategory.ID))
	}

	// Chuyển đổi số tiền
	amount, err := strconv.Atoi(request.Amount)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// Xử lý ngày thu nhập
	var incomeDate time.Time
	if request.IncomeDate != "" {
		// Parse ngày từ form (format YYYY-MM-DD)
		parsedDate, err := time.Parse("2006-01-02", request.IncomeDate)
		if err == nil {
			incomeDate = parsedDate
		} else {
			// Nếu không parse được, sử dụng ngày hiện tại
			incomeDate = time.Now()
		}
	} else {
		// Nếu không có ngày được chọn, sử dụng ngày hiện tại
		incomeDate = time.Now()
	}

	// Tạo khoản thu nhập
	income := models.Income{
		CategoryID: uint(categoryID),
		UserID:     userID.(uint),
		Amount:     amount,
		Note:       request.Note,
		IncomeDate: incomeDate,
	}

	result := database.DB.Create(&income)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.Redirect(http.StatusSeeOther, "/income")
}

// GetIncomeCategories trả về danh sách các danh mục thu nhập
func GetIncomeCategories(c *gin.Context) {
	var categories []models.IncomeCategory
	database.DB.Find(&categories)
	c.JSON(http.StatusOK, categories)
}

// AddIncomeCategory xử lý việc thêm danh mục thu nhập mới
func AddIncomeCategory(c *gin.Context) {
	var request struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Kiểm tra tên danh mục không được để trống
	if request.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tên danh mục không được để trống"})
		return
	}

	// Tạo danh mục mới
	newCategory := models.IncomeCategory{Name: request.Name}

	// Kiểm tra xem danh mục đã tồn tại chưa
	var existingCategory models.IncomeCategory
	result := database.DB.Where("name = ?", request.Name).First(&existingCategory)

	if result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Danh mục này đã tồn tại"})
		return
	}

	// Lưu danh mục mới vào database
	result = database.DB.Create(&newCategory)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Thêm danh mục thu nhập thành công",
		"category": newCategory,
	})
}

// GetIncomes trả về danh sách các khoản thu nhập
func GetIncomes(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var incomes []models.Income
	database.DB.Preload("Category").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&incomes)

	c.JSON(http.StatusOK, incomes)
}

// DeleteIncome xử lý việc xóa khoản thu nhập
func DeleteIncome(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	// Kiểm tra ID hợp lệ
	incomeID, err := strconv.Atoi(id)
	if err != nil || incomeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Kiểm tra khoản thu nhập thuộc về người dùng hiện tại
	var income models.Income
	if err := database.DB.Where("id = ? AND user_id = ?", incomeID, userID).First(&income).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy khoản thu nhập hoặc bạn không có quyền xóa"})
		return
	}

	// Thực hiện xóa
	result := database.DB.Delete(&income)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa thành công"})
}

// GetIncomeSummary trả về tổng thu nhập theo ngày và tháng
func GetIncomeSummary(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var (
		dailyTotal   int
		monthlyTotal int
		yearlyTotal  int
	)

	now := time.Now()

	// Tính tổng ngày theo ngày thu nhập
	database.DB.Table("incomes").
		Where("DATE(income_date) = ? AND user_id = ? AND deleted_at IS NULL", now.Format("2006-01-02"), userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&dailyTotal)

	// Tính tổng tháng theo ngày thu nhập
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	database.DB.Table("incomes").
		Where("income_date >= ? AND income_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfMonth, firstOfMonth.AddDate(0, 1, 0), userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyTotal)

	// Tính tổng năm theo ngày thu nhập
	firstOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	database.DB.Table("incomes").
		Where("income_date >= ? AND income_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfYear, firstOfYear.AddDate(1, 0, 0), userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&yearlyTotal)

	c.JSON(http.StatusOK, gin.H{
		"daily":   dailyTotal,
		"monthly": monthlyTotal,
		"yearly":  yearlyTotal,
	})
}

// GetMonthlyIncomes trả về thu nhập theo tháng trong năm hiện tại
func GetMonthlyIncomes(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	now := time.Now()
	currentYear := now.Year()

	// Lấy ngày đầu tiên của năm
	firstOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, now.Location())

	// Lấy ngày đầu tiên của năm tiếp theo
	firstOfNextYear := time.Date(currentYear+1, 1, 1, 0, 0, 0, 0, now.Location())

	// Tạo mảng kết quả với 12 tháng
	result := make([]int, 12)

	// Truy vấn tổng thu nhập theo tháng
	type MonthlySum struct {
		Month int
		Total int
	}

	var monthlySums []MonthlySum

	// Truy vấn SQL để lấy tổng thu nhập theo tháng
	database.DB.Table("incomes").
		Select("EXTRACT(MONTH FROM income_date) as month, COALESCE(SUM(amount), 0) as total").
		Where("income_date >= ? AND income_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfYear, firstOfNextYear, userID).
		Group("EXTRACT(MONTH FROM income_date)").
		Scan(&monthlySums)

	// Khởi tạo mảng kết quả với giá trị 0
	for i := range result {
		result[i] = 0
	}

	// Điền dữ liệu vào mảng kết quả
	for _, sum := range monthlySums {
		if sum.Month >= 1 && sum.Month <= 12 {
			result[sum.Month-1] = sum.Total
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"year": currentYear,
	})
}

// GetFinancialOverview trả về tổng quan tài chính (thu nhập, chi tiêu, tiết kiệm)
func GetFinancialOverview(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var (
		monthlyIncome  int
		monthlyExpense int
		balance        int
		savingTotal    int
	)

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := firstOfMonth.AddDate(0, 1, 0)

	// Tính tổng thu nhập tháng
	database.DB.Table("incomes").
		Where("income_date >= ? AND income_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfMonth, nextMonth, userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyIncome)

	// Tính tổng chi tiêu tháng
	database.DB.Table("expenses").
		Where("expense_date >= ? AND expense_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfMonth, nextMonth, userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyExpense)

	// Tính tổng tiết kiệm
	database.DB.Table("saving_goals").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("COALESCE(SUM(current_amount), 0)").
		Scan(&savingTotal)

	// Tính số dư (thu nhập - chi tiêu)
	balance = monthlyIncome - monthlyExpense

	c.JSON(http.StatusOK, gin.H{
		"monthlyIncome":  monthlyIncome,
		"monthlyExpense": monthlyExpense,
		"balance":        balance,
		"savingTotal":    savingTotal,
	})
}

// ShowIncomePage hiển thị trang quản lý thu nhập
func ShowIncomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "income.html", nil)
}