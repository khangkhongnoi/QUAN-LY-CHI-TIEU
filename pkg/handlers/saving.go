package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// AddSavingGoal xử lý thêm mục tiêu tiết kiệm mới
func AddSavingGoal(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		Name         string `json:"name" binding:"required"`
		TargetAmount string `json:"target_amount" binding:"required"`
		Deadline     string `json:"deadline"`
		Description  string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chuyển đổi số tiền mục tiêu
	targetAmount, err := strconv.Atoi(request.TargetAmount)
	if err != nil || targetAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số tiền mục tiêu không hợp lệ"})
		return
	}

	// Xử lý thời hạn
	var deadline time.Time
	if request.Deadline != "" {
		parsedDeadline, err := time.Parse("2006-01-02", request.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thời hạn không hợp lệ"})
			return
		}
		deadline = parsedDeadline
	}

	// Tạo mục tiêu tiết kiệm mới
	savingGoal := models.SavingGoal{
		UserID:       userID.(uint),
		Name:         request.Name,
		TargetAmount: targetAmount,
		Deadline:     deadline,
		Description:  request.Description,
	}

	result := database.DB.Create(&savingGoal)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tạo mục tiêu tiết kiệm thành công",
		"goal":    savingGoal,
	})
}

// GetSavingGoals trả về danh sách mục tiêu tiết kiệm
func GetSavingGoals(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var goals []models.SavingGoal
	database.DB.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&goals)

	c.JSON(http.StatusOK, goals)
}

// AddSavingTransaction xử lý thêm giao dịch tiết kiệm
func AddSavingTransaction(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		GoalID   string `json:"goal_id" binding:"required"`
		Amount   string `json:"amount" binding:"required"`
		Date     string `json:"date"`
		Note     string `json:"note"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chuyển đổi goal_id
	goalID, err := strconv.ParseUint(request.GoalID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mục tiêu không hợp lệ"})
		return
	}

	// Kiểm tra mục tiêu tồn tại và thuộc về người dùng
	var goal models.SavingGoal
	if err := database.DB.Where("id = ? AND user_id = ?", goalID, userID).First(&goal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy mục tiêu tiết kiệm"})
		return
	}

	// Chuyển đổi số tiền
	amount, err := strconv.Atoi(request.Amount)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số tiền không hợp lệ"})
		return
	}

	// Xử lý ngày giao dịch
	var transactionDate time.Time
	if request.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", request.Date)
		if err == nil {
			transactionDate = parsedDate
		} else {
			transactionDate = time.Now()
		}
	} else {
		transactionDate = time.Now()
	}

	// Tạo giao dịch tiết kiệm mới
	transaction := models.SavingTransaction{
		UserID:          userID.(uint),
		GoalID:          uint(goalID),
		Amount:          amount,
		TransactionDate: transactionDate,
		Note:            request.Note,
	}

	// Bắt đầu transaction để đảm bảo tính nhất quán dữ liệu
	tx := database.DB.Begin()

	// Tạo giao dịch
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Cập nhật số tiền hiện tại của mục tiêu
	goal.CurrentAmount += amount
	
	// Kiểm tra nếu đã đạt mục tiêu
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.Completed = true
	}

	if err := tx.Save(&goal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":     "Thêm giao dịch tiết kiệm thành công",
		"transaction": transaction,
		"goal":        goal,
	})
}

// GetSavingTransactions trả về danh sách giao dịch tiết kiệm cho một mục tiêu
func GetSavingTransactions(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	goalID := c.Param("goal_id")

	// Kiểm tra goal_id hợp lệ
	goalIDInt, err := strconv.Atoi(goalID)
	if err != nil || goalIDInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID mục tiêu không hợp lệ"})
		return
	}

	// Kiểm tra mục tiêu tồn tại và thuộc về người dùng
	var goal models.SavingGoal
	if err := database.DB.Where("id = ? AND user_id = ?", goalIDInt, userID).First(&goal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy mục tiêu tiết kiệm"})
		return
	}

	// Lấy danh sách giao dịch
	var transactions []models.SavingTransaction
	database.DB.Where("goal_id = ? AND user_id = ? AND deleted_at IS NULL", goalIDInt, userID).
		Order("transaction_date DESC").
		Find(&transactions)

	c.JSON(http.StatusOK, transactions)
}

// DeleteSavingGoal xử lý việc xóa mục tiêu tiết kiệm
func DeleteSavingGoal(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	// Kiểm tra ID hợp lệ
	goalID, err := strconv.Atoi(id)
	if err != nil || goalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Kiểm tra mục tiêu thuộc về người dùng hiện tại
	var goal models.SavingGoal
	if err := database.DB.Where("id = ? AND user_id = ?", goalID, userID).First(&goal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy mục tiêu tiết kiệm hoặc bạn không có quyền xóa"})
		return
	}

	// Bắt đầu transaction để xóa cả mục tiêu và các giao dịch liên quan
	tx := database.DB.Begin()

	// Xóa các giao dịch liên quan
	if err := tx.Where("goal_id = ?", goalID).Delete(&models.SavingTransaction{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Xóa mục tiêu
	if err := tx.Delete(&goal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Commit transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Xóa mục tiêu tiết kiệm thành công"})
}

// ShowSavingPage hiển thị trang quản lý tiết kiệm
func ShowSavingPage(c *gin.Context) {
	c.HTML(http.StatusOK, "saving.html", nil)
}