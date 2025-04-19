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
	var request struct {
		CategoryID  string `form:"category_id" binding:"required"`
		NewCategory string `form:"new_category"`
		Amount      string `form:"amount" binding:"required"`
		Note        string `form:"note"`
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

	// Tạo expense
	expense := models.Expense{
		CategoryID: uint(categoryID),
		Amount:     amount,
		Note:       request.Note,
	}

	// Xử lý file hình ảnh nếu có
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()

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

		// Nếu lưu file thất bại, thử lưu dưới dạng base64
		if err != nil {
			// Đọc lại file từ đầu
			file.Seek(0, 0)

			// Đọc nội dung file
			fileBytes, err := io.ReadAll(file)
			if err == nil {
				// Mã hóa base64
				contentType := http.DetectContentType(fileBytes)
				base64Data := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(fileBytes))

				// Lưu dữ liệu base64
				expense.ImageData = base64Data
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
	var (
		dailyTotal   int
		monthlyTotal int
	)

	now := time.Now()

	// Tính tổng ngày
	database.DB.Model(&models.Expense{}).
		Where("DATE(created_at) = ?", now.Format("2006-01-02")).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&dailyTotal)

	// Tính tổng tháng
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	database.DB.Model(&models.Expense{}).
		Where("created_at >= ? AND created_at < ?", firstOfMonth, firstOfMonth.AddDate(0, 1, 0)).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyTotal)

	c.JSON(http.StatusOK, gin.H{
		"daily":   dailyTotal,
		"monthly": monthlyTotal,
	})
}

func GetExpenses(c *gin.Context) {
	var expenses []models.Expense
	database.DB.Preload("Category").
		Order("created_at DESC").
		Find(&expenses)

	c.JSON(http.StatusOK, expenses)
}

func GetCategories(c *gin.Context) {
	var categories []models.Category
	database.DB.Find(&categories)
	c.JSON(http.StatusOK, categories)
}
func DeleteExpense(c *gin.Context) {
	id := c.Param("id")

	// Kiểm tra ID hợp lệ
	expenseID, err := strconv.Atoi(id)
	if err != nil || expenseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	// Thực hiện xóa
	result := database.DB.Delete(&models.Expense{}, expenseID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Kiểm tra có bản ghi nào bị xóa không
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy chi tiêu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa thành công"})
}

// GetDailyExpenses trả về chi tiêu theo ngày trong tháng hiện tại
func GetDailyExpenses(c *gin.Context) {
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
	
	// Truy vấn SQL để lấy tổng chi tiêu theo ngày
	database.DB.Model(&models.Expense{}).
		Select("EXTRACT(DAY FROM created_at) as day, COALESCE(SUM(amount), 0) as total").
		Where("created_at >= ? AND created_at < ?", firstOfMonth, firstOfNextMonth).
		Group("EXTRACT(DAY FROM created_at)").
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
