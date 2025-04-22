package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// AddBudget xử lý thêm kế hoạch ngân sách mới
func AddBudget(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		CategoryID  string `json:"category_id" binding:"required"`
		Amount      string `json:"amount" binding:"required"`
		StartDate   string `json:"start_date" binding:"required"`
		EndDate     string `json:"end_date" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chuyển đổi category_id
	categoryID, err := strconv.ParseUint(request.CategoryID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Danh mục không hợp lệ"})
		return
	}

	// Chuyển đổi số tiền
	amount, err := strconv.Atoi(request.Amount)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số tiền không hợp lệ"})
		return
	}

	// Chuyển đổi ngày bắt đầu
	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày bắt đầu không hợp lệ"})
		return
	}

	// Chuyển đổi ngày kết thúc
	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày kết thúc không hợp lệ"})
		return
	}

	// Kiểm tra ngày kết thúc phải sau ngày bắt đầu
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày kết thúc phải sau ngày bắt đầu"})
		return
	}

	// Tạo kế hoạch ngân sách mới
	budget := models.Budget{
		UserID:      userID.(uint),
		CategoryID:  uint(categoryID),
		Amount:      amount,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: request.Description,
	}

	result := database.DB.Create(&budget)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tạo kế hoạch ngân sách thành công",
		"budget":  budget,
	})
}

// GetBudgets trả về danh sách kế hoạch ngân sách
func GetBudgets(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var budgets []models.Budget
	database.DB.Preload("Category").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&budgets)

	c.JSON(http.StatusOK, budgets)
}

// GetBudgetSummary trả về tổng hợp tình hình ngân sách
func GetBudgetSummary(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy tất cả ngân sách hiện tại (chưa kết thúc)
	var budgets []models.Budget
	database.DB.Preload("Category").
		Where("user_id = ? AND end_date >= ? AND deleted_at IS NULL", userID, time.Now()).
		Find(&budgets)

	// Tạo mảng kết quả
	var summaries []models.BudgetSummary

	// Tính toán chi tiết cho từng ngân sách
	for _, budget := range budgets {
		var spentAmount int

		// Tính tổng chi tiêu trong khoảng thời gian của ngân sách
		database.DB.Table("expenses").
			Where("user_id = ? AND category_id = ? AND expense_date >= ? AND expense_date <= ? AND deleted_at IS NULL",
				userID, budget.CategoryID, budget.StartDate, budget.EndDate).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&spentAmount)

		// Tính số tiền còn lại và phần trăm đã sử dụng
		remainingAmount := budget.Amount - spentAmount
		var percentUsed float64 = 0
		if budget.Amount > 0 {
			percentUsed = float64(spentAmount) / float64(budget.Amount) * 100
		}

		// Tạo đối tượng tổng hợp
		summary := models.BudgetSummary{
			BudgetID:        budget.ID,
			CategoryID:      budget.CategoryID,
			CategoryName:    budget.Category.Name,
			BudgetAmount:    budget.Amount,
			SpentAmount:     spentAmount,
			RemainingAmount: remainingAmount,
			PercentUsed:     percentUsed,
		}

		summaries = append(summaries, summary)
	}

	c.JSON(http.StatusOK, summaries)
}

// DeleteBudget xử lý việc xóa kế hoạch ngân sách
func DeleteBudget(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	// Kiểm tra ID hợp lệ
	budgetID, err := strconv.Atoi(id)
	if err != nil || budgetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Kiểm tra ngân sách thuộc về người dùng hiện tại
	var budget models.Budget
	if err := database.DB.Where("id = ? AND user_id = ?", budgetID, userID).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy kế hoạch ngân sách hoặc bạn không có quyền xóa"})
		return
	}

	// Thực hiện xóa
	result := database.DB.Delete(&budget)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa kế hoạch ngân sách thành công"})
}

// ShowBudgetPage hiển thị trang quản lý ngân sách
func ShowBudgetPage(c *gin.Context) {
	c.HTML(http.StatusOK, "budget.html", nil)
}