package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func AddExpense(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		CategoryID  string `form:"category_id" binding:"required"`
		NewCategory string `form:"new_category"`
		Amount      string `form:"amount" binding:"required"`
		Note        string `form:"note"`
		ExpenseDate string `form:"expense_date"`
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

	if categoryID == 5 && request.NewCategory != "" { // ID 5 = Đồ ăn khác
		newCategory := models.Category{Name: request.NewCategory}
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

	// Xử lý ngày chi tiêu
	var expenseDate time.Time
	if request.ExpenseDate != "" {
		// Parse ngày từ form (format YYYY-MM-DD)
		parsedDate, err := time.Parse("2006-01-02", request.ExpenseDate)
		if err == nil {
			expenseDate = parsedDate
		} else {
			// Nếu không parse được, sử dụng ngày hiện tại
			expenseDate = time.Now()
		}
	} else {
		// Nếu không có ngày được chọn, sử dụng ngày hiện tại
		expenseDate = time.Now()
	}

	// Tạo expense
	expense := models.Expense{
		CategoryID:  uint(categoryID),
		UserID:      userID.(uint), // Thêm user_id vào expense
		Amount:      amount,
		Note:        request.Note,
		ExpenseDate: expenseDate,
	}

	// Xử lý file hình ảnh nếu có
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()

		// Đọc nội dung file để lưu dưới dạng base64 (luôn lưu để dự phòng)
		fileBytes, err := io.ReadAll(file)
		if err == nil {
			// Mã hóa base64
			contentType := http.DetectContentType(fileBytes)
			base64Data := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(fileBytes))
			expense.ImageData = base64Data
		}

		// Reset con trỏ file để có thể đọc lại từ đầu
		file.Seek(0, 0)

		// Tạo thư mục uploads nếu chưa tồn tại
		uploadsDir := "./static/uploads"
		if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
			os.MkdirAll(uploadsDir, 0755)
		}

		// Tạo tên file duy nhất
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
		filepath := filepath.Join(uploadsDir, filename)

		// Lưu file vào thư mục uploads
		out, err := os.Create(filepath)
		if err == nil {
			defer out.Close()
			_, err = io.Copy(out, file)
			if err == nil {
				// Lưu đường dẫn file
				expense.ImagePath = "/static/uploads/" + filename
			}
		}
	}

	result := database.DB.Create(&expense)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func GetSummary(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var (
		dailyTotal   int
		monthlyTotal int
	)

	now := time.Now()

	// Tính tổng ngày theo ngày chi tiêu
	database.DB.Table("expenses").
		Where("DATE(expense_date) = ? AND user_id = ? AND deleted_at IS NULL", now.Format("2006-01-02"), userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&dailyTotal)

	// Tính tổng tháng theo ngày chi tiêu
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	database.DB.Table("expenses").
		Where("expense_date >= ? AND expense_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfMonth, firstOfMonth.AddDate(0, 1, 0), userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyTotal)

	c.JSON(http.StatusOK, gin.H{
		"daily":   dailyTotal,
		"monthly": monthlyTotal,
	})
}

func GetExpenses(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var expenses []models.Expense
	database.DB.Preload("Category").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&expenses)

	c.JSON(http.StatusOK, expenses)
}

func GetCategories(c *gin.Context) {
	var categories []models.Category
	database.DB.Find(&categories)
	c.JSON(http.StatusOK, categories)
}

// AddCategory xử lý việc thêm danh mục mới
func AddCategory(c *gin.Context) {
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
	newCategory := models.Category{Name: request.Name}

	// Kiểm tra xem danh mục đã tồn tại chưa
	var existingCategory models.Category
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
		"message":  "Thêm danh mục thành công",
		"category": newCategory,
	})
}
func DeleteExpense(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	// Kiểm tra ID hợp lệ
	expenseID, err := strconv.Atoi(id)
	if err != nil || expenseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Kiểm tra chi tiêu thuộc về người dùng hiện tại
	var expense models.Expense
	if err := database.DB.Where("id = ? AND user_id = ?", expenseID, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chi tiêu hoặc bạn không có quyền xóa"})
		return
	}

	// Thực hiện xóa
	result := database.DB.Delete(&expense)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa thành công"})
}

// GetExpenseDetail trả về chi tiết của một khoản chi tiêu
func GetExpenseDetail(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	// Kiểm tra ID hợp lệ
	expenseID, err := strconv.Atoi(id)
	if err != nil || expenseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Lấy chi tiết chi tiêu
	var expense models.Expense
	if err := database.DB.Preload("Category").Where("id = ? AND user_id = ?", expenseID, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chi tiêu"})
		return
	}

	c.JSON(http.StatusOK, expense)
}

// UpdateExpense xử lý việc cập nhật chi tiêu
func UpdateExpense(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	// Kiểm tra ID hợp lệ
	expenseID, err := strconv.Atoi(id)
	if err != nil || expenseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Kiểm tra chi tiêu tồn tại và thuộc về người dùng
	var expense models.Expense
	if err := database.DB.Where("id = ? AND user_id = ?", expenseID, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chi tiêu hoặc bạn không có quyền chỉnh sửa"})
		return
	}

	var request struct {
		CategoryID  string `form:"category_id"`
		Amount      string `form:"amount"`
		Note        string `form:"note"`
		ExpenseDate string `form:"expense_date"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cập nhật danh mục nếu có
	if request.CategoryID != "" {
		categoryID, err := strconv.ParseUint(request.CategoryID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Danh mục không hợp lệ"})
			return
		}
		expense.CategoryID = uint(categoryID)
	}

	// Cập nhật số tiền nếu có
	if request.Amount != "" {
		amount, err := strconv.Atoi(request.Amount)
		if err != nil || amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số tiền không hợp lệ"})
			return
		}
		expense.Amount = amount
	}

	// Cập nhật ghi chú
	if request.Note != "" {
		expense.Note = request.Note
	}

	// Cập nhật ngày chi tiêu nếu có
	if request.ExpenseDate != "" {
		expenseDate, err := time.Parse("2006-01-02", request.ExpenseDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày chi tiêu không hợp lệ"})
			return
		}
		expense.ExpenseDate = expenseDate
	}

	// Xử lý file hình ảnh mới nếu có
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()

		// Đọc nội dung file để lưu dưới dạng base64
		fileBytes, err := io.ReadAll(file)
		if err == nil {
			// Mã hóa base64
			contentType := http.DetectContentType(fileBytes)
			base64Data := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(fileBytes))
			expense.ImageData = base64Data
		}

		// Reset con trỏ file để có thể đọc lại từ đầu
		file.Seek(0, 0)

		// Tạo thư mục uploads nếu chưa tồn tại
		uploadsDir := "./static/uploads"
		if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
			os.MkdirAll(uploadsDir, 0755)
		}

		// Tạo tên file duy nhất
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
		filepath := filepath.Join(uploadsDir, filename)

		// Lưu file vào thư mục uploads
		out, err := os.Create(filepath)
		if err == nil {
			defer out.Close()
			_, err = io.Copy(out, file)
			if err == nil {
				// Lưu đường dẫn file
				expense.ImagePath = "/static/uploads/" + filename
			}
		}
	}

	// Lưu các thay đổi
	result := database.DB.Save(&expense)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật chi tiêu thành công",
		"expense": expense,
	})
}

// GetDailyExpenses trả về chi tiêu theo ngày trong tháng hiện tại
func GetDailyExpenses(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	now := time.Now()

	// Lấy ngày đầu tiên của tháng
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Lấy ngày đầu tiên của tháng tiếp theo
	firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)

	// Số ngày trong tháng
	daysInMonth := firstOfNextMonth.Add(-time.Hour).Day()

	// Tạo mảng kết quả với số ngày trong tháng
	result := make([]int, daysInMonth)

	// Truy vấn tổng chi tiêu theo ngày
	type DailySum struct {
		Day   int
		Total int
	}

	var dailySums []DailySum

	// Truy vấn SQL để lấy tổng chi tiêu theo ngày chi tiêu
	database.DB.Table("expenses").
		Select("EXTRACT(DAY FROM expense_date) as day, COALESCE(SUM(amount), 0) as total").
		Where("expense_date >= ? AND expense_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfMonth, firstOfNextMonth, userID).
		Group("EXTRACT(DAY FROM expense_date)").
		Scan(&dailySums)

	// Khởi tạo mảng kết quả với giá trị 0
	for i := range result {
		result[i] = 0
	}

	// Điền dữ liệu vào mảng kết quả
	for _, sum := range dailySums {
		if sum.Day >= 1 && sum.Day <= daysInMonth {
			result[sum.Day-1] = sum.Total
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"days": daysInMonth,
	})
}

// GetMonthlyExpenses trả về chi tiêu theo tháng trong năm hiện tại
func GetMonthlyExpenses(c *gin.Context) {
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

	// Truy vấn tổng chi tiêu theo tháng
	type MonthlySum struct {
		Month int
		Total int
	}

	var monthlySums []MonthlySum

	// Truy vấn SQL để lấy tổng chi tiêu theo tháng chi tiêu
	database.DB.Table("expenses").
		Select("EXTRACT(MONTH FROM expense_date) as month, COALESCE(SUM(amount), 0) as total").
		Where("expense_date >= ? AND expense_date < ? AND user_id = ? AND deleted_at IS NULL",
			firstOfYear, firstOfNextYear, userID).
		Group("EXTRACT(MONTH FROM expense_date)").
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
