package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// GetBudgetWarnings trả về các cảnh báo ngân sách
func GetBudgetWarnings(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy tất cả ngân sách của người dùng
	var budgets []models.Budget
	database.DB.Where("user_id = ?", userID).Find(&budgets)

	var warnings []gin.H

	for _, budget := range budgets {
		// Tính tổng chi tiêu trong danh mục này trong tháng hiện tại
		var spent int
		now := time.Now()
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		database.DB.Table("expenses").
			Where("category_id = ? AND user_id = ? AND expense_date >= ? AND expense_date < ? AND deleted_at IS NULL",
				budget.CategoryID, userID, firstOfMonth, firstOfMonth.AddDate(0, 1, 0)).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&spent)

		// Tính phần trăm đã chi tiêu
		percentSpent := float64(spent) / float64(budget.Amount) * 100

		// Lấy tên danh mục
		var category models.Category
		database.DB.First(&category, budget.CategoryID)

		// Tạo cảnh báo dựa trên mức độ chi tiêu
		if percentSpent >= 90 {
			warnings = append(warnings, gin.H{
				"category_id":   budget.CategoryID,
				"category_name": category.Name,
				"budget":        budget.Amount,
				"spent":         spent,
				"percent":       percentSpent,
				"severity":      "high",
				"message":       "Bạn đã chi tiêu gần hết ngân sách cho danh mục này!",
			})

			// Lưu cảnh báo vào database nếu chưa có
			var existingWarning models.BudgetWarning
			result := database.DB.Where("user_id = ? AND budget_id = ? AND warning_type = ? AND dismissed = ? AND created_at >= ?",
				userID, budget.ID, "approaching_limit", false, firstOfMonth).First(&existingWarning)

			if result.RowsAffected == 0 {
				warning := models.BudgetWarning{
					UserID:      userID.(uint),
					BudgetID:    budget.ID,
					CategoryID:  budget.CategoryID,
					WarningType: "approaching_limit",
					Threshold:   90,
					Message:     "Bạn đã chi tiêu gần hết ngân sách cho danh mục " + category.Name,
					Dismissed:   false,
				}
				database.DB.Create(&warning)
			}
		} else if percentSpent >= 75 {
			warnings = append(warnings, gin.H{
				"category_id":   budget.CategoryID,
				"category_name": category.Name,
				"budget":        budget.Amount,
				"spent":         spent,
				"percent":       percentSpent,
				"severity":      "medium",
				"message":       "Bạn đã chi tiêu hơn 75% ngân sách cho danh mục này.",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"warnings": warnings,
	})
}

// DetectUnusualExpenses phát hiện chi tiêu bất thường
func DetectUnusualExpenses(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy chi tiêu trong 3 tháng gần nhất để tính trung bình
	now := time.Now()
	threeMonthsAgo := now.AddDate(0, -3, 0)

	type CategoryAvg struct {
		CategoryID uint
		Name       string
		AvgAmount  float64
	}

	var categoryAvgs []CategoryAvg

	// Tính chi tiêu trung bình theo danh mục trong 3 tháng qua
	database.DB.Table("expenses").
		Select("expenses.category_id, categories.name, AVG(expenses.amount) as avg_amount").
		Joins("JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ? AND expenses.expense_date >= ? AND expenses.deleted_at IS NULL",
			userID, threeMonthsAgo).
		Group("expenses.category_id, categories.name").
		Scan(&categoryAvgs)

	// Lấy chi tiêu tháng hiện tại theo danh mục
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type CategoryCurrent struct {
		CategoryID  uint
		Name        string
		TotalAmount int
	}

	var categoryCurrent []CategoryCurrent

	database.DB.Table("expenses").
		Select("expenses.category_id, categories.name, SUM(expenses.amount) as total_amount").
		Joins("JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ? AND expenses.expense_date >= ? AND expenses.expense_date < ? AND expenses.deleted_at IS NULL",
			userID, firstOfMonth, firstOfMonth.AddDate(0, 1, 0)).
		Group("expenses.category_id, categories.name").
		Scan(&categoryCurrent)

	// So sánh và phát hiện chi tiêu bất thường
	var unusualExpenses []gin.H

	for _, current := range categoryCurrent {
		for _, avg := range categoryAvgs {
			if current.CategoryID == avg.CategoryID {
				// Nếu chi tiêu hiện tại vượt quá 150% trung bình
				if float64(current.TotalAmount) > avg.AvgAmount*1.5 && avg.AvgAmount > 0 {
					percentIncrease := (float64(current.TotalAmount) - avg.AvgAmount) / avg.AvgAmount * 100
					unusualExpenses = append(unusualExpenses, gin.H{
						"category_id":      current.CategoryID,
						"category_name":    current.Name,
						"average_amount":   avg.AvgAmount,
						"current_amount":   current.TotalAmount,
						"percent_increase": percentIncrease,
						"message":          "Chi tiêu cao bất thường so với trung bình 3 tháng qua",
					})
				}
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"unusual_expenses": unusualExpenses,
	})
}

// ForecastExpenses dự báo chi tiêu trong tương lai
func ForecastExpenses(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy dữ liệu chi tiêu trong 6 tháng gần nhất
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)

	type MonthlyExpense struct {
		Month      int
		Year       int
		CategoryID uint
		Amount     int
	}

	var monthlyExpenses []MonthlyExpense

	// Truy vấn chi tiêu theo tháng và danh mục
	database.DB.Table("expenses").
		Select("EXTRACT(MONTH FROM expense_date) as month, EXTRACT(YEAR FROM expense_date) as year, category_id, SUM(amount) as amount").
		Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL", userID, sixMonthsAgo).
		Group("EXTRACT(MONTH FROM expense_date), EXTRACT(YEAR FROM expense_date), category_id").
		Scan(&monthlyExpenses)

	// Tổ chức dữ liệu theo danh mục
	categoryExpenses := make(map[uint][]int)
	for _, expense := range monthlyExpenses {
		if _, exists := categoryExpenses[expense.CategoryID]; !exists {
			categoryExpenses[expense.CategoryID] = make([]int, 0)
		}
		categoryExpenses[expense.CategoryID] = append(categoryExpenses[expense.CategoryID], expense.Amount)
	}

	// Dự báo chi tiêu cho tháng tiếp theo
	nextMonth := now.Month() + 1
	nextYear := now.Year()
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}

	var forecasts []gin.H
	var categories []models.Category
	database.DB.Find(&categories)

	categoryMap := make(map[uint]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}

	for catID, expenses := range categoryExpenses {
		if len(expenses) < 3 {
			continue // Bỏ qua nếu không đủ dữ liệu
		}

		// Tính trung bình đơn giản
		var sum int
		for _, amount := range expenses {
			sum += amount
		}
		avgAmount := sum / len(expenses)

		// Tính độ tin cậy dựa trên độ lệch chuẩn
		var variance float64
		for _, amount := range expenses {
			variance += math.Pow(float64(amount-avgAmount), 2)
		}
		variance /= float64(len(expenses))
		stdDev := math.Sqrt(variance)

		// Độ tin cậy giảm khi độ lệch chuẩn tăng
		confidence := 1.0
		if avgAmount > 0 {
			// Tính độ tin cậy dựa trên tỷ lệ độ lệch chuẩn / trung bình
			// Độ tin cậy giảm khi độ lệch chuẩn tăng
			confidence = math.Max(0.1, math.Min(1.0, 1.0-(stdDev/float64(avgAmount))/2))
		}

		// Lưu dự báo vào database
		forecast := models.ExpenseForecast{
			UserID:     userID.(uint),
			CategoryID: catID,
			Month:      int(nextMonth),
			Year:       nextYear,
			Amount:     avgAmount,
			Confidence: confidence,
		}

		// Kiểm tra xem đã có dự báo cho tháng này chưa
		var existingForecast models.ExpenseForecast
		result := database.DB.Where("user_id = ? AND category_id = ? AND month = ? AND year = ?",
			userID, catID, nextMonth, nextYear).First(&existingForecast)

		if result.RowsAffected == 0 {
			database.DB.Create(&forecast)
		} else {
			existingForecast.Amount = avgAmount
			existingForecast.Confidence = confidence
			database.DB.Save(&existingForecast)
		}

		// Thêm vào kết quả trả về
		forecasts = append(forecasts, gin.H{
			"category_id":   catID,
			"category_name": categoryMap[catID],
			"month":         int(nextMonth),
			"year":          nextYear,
			"amount":        avgAmount,
			"confidence":    confidence,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"forecasts": forecasts,
	})
}

// GetSavingChallenges trả về các thách thức tiết kiệm
func GetSavingChallenges(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var challenges []models.SavingChallenge
	database.DB.Where("user_id = ?", userID).Find(&challenges)

	// Thêm thông tin danh mục
	var result []gin.H
	for _, challenge := range challenges {
		var category models.Category
		database.DB.First(&category, challenge.CategoryID)

		// Tính phần trăm hoàn thành
		percentComplete := 0.0
		if challenge.TargetAmount > 0 {
			percentComplete = float64(challenge.CurrentAmount) / float64(challenge.TargetAmount) * 100
		}

		result = append(result, gin.H{
			"id":               challenge.ID,
			"title":            challenge.Title,
			"description":      challenge.Description,
			"target_amount":    challenge.TargetAmount,
			"current_amount":   challenge.CurrentAmount,
			"start_date":       challenge.StartDate,
			"end_date":         challenge.EndDate,
			"category_name":    category.Name,
			"status":           challenge.Status,
			"percent_complete": percentComplete,
			"days_remaining":   int(challenge.EndDate.Sub(time.Now()).Hours() / 24),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"challenges": result,
	})
}

// CreateSavingChallenge tạo một thách thức tiết kiệm mới
func CreateSavingChallenge(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		Title        string `json:"title" binding:"required"`
		Description  string `json:"description"`
		TargetAmount int    `json:"target_amount" binding:"required"`
		CategoryID   uint   `json:"category_id" binding:"required"`
		EndDate      string `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Kiểm tra số tiền mục tiêu
	if request.TargetAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Số tiền mục tiêu phải lớn hơn 0"})
		return
	}

	// Kiểm tra ngày kết thúc
	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày kết thúc không hợp lệ"})
		return
	}

	if endDate.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ngày kết thúc phải sau ngày hiện tại"})
		return
	}

	// Tạo thách thức mới
	challenge := models.SavingChallenge{
		UserID:        userID.(uint),
		Title:         request.Title,
		Description:   request.Description,
		TargetAmount:  request.TargetAmount,
		CurrentAmount: 0,
		StartDate:     time.Now(),
		EndDate:       endDate,
		CategoryID:    request.CategoryID,
		Status:        "active",
	}

	result := database.DB.Create(&challenge)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Tạo thách thức tiết kiệm thành công",
		"challenge": challenge,
	})
}

// AnalyzeExpensePatterns phân tích mẫu chi tiêu
func AnalyzeExpensePatterns(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy dữ liệu chi tiêu trong 6 tháng gần nhất
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)

	var expenses []models.Expense
	database.DB.Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL", userID, sixMonthsAgo).
		Preload("Category").Find(&expenses)

	// Phân tích chi tiêu định kỳ
	recurringExpenses := analyzeRecurringExpenses(expenses)

	// Phân tích chi tiêu theo mùa
	seasonalExpenses := analyzeSeasonalExpenses(expenses)

	// Phân tích chi tiêu bốc đồng
	impulseExpenses := analyzeImpulseExpenses(expenses)

	// Lưu các mẫu vào database
	saveExpensePatterns(userID.(uint), recurringExpenses, "recurring")
	saveExpensePatterns(userID.(uint), seasonalExpenses, "seasonal")
	saveExpensePatterns(userID.(uint), impulseExpenses, "impulse")

	c.JSON(http.StatusOK, gin.H{
		"recurring_expenses": recurringExpenses,
		"seasonal_expenses":  seasonalExpenses,
		"impulse_expenses":   impulseExpenses,
	})
}

// Hàm phân tích chi tiêu định kỳ
func analyzeRecurringExpenses(expenses []models.Expense) []gin.H {
	// Nhóm chi tiêu theo danh mục
	categoryExpenses := make(map[uint][]models.Expense)
	for _, expense := range expenses {
		if _, exists := categoryExpenses[expense.CategoryID]; !exists {
			categoryExpenses[expense.CategoryID] = make([]models.Expense, 0)
		}
		categoryExpenses[expense.CategoryID] = append(categoryExpenses[expense.CategoryID], expense)
	}

	var recurringExpenses []gin.H

	for catID, catExpenses := range categoryExpenses {
		if len(catExpenses) < 3 {
			continue // Bỏ qua nếu không đủ dữ liệu
		}

		// Sắp xếp theo ngày
		sort.Slice(catExpenses, func(i, j int) bool {
			return catExpenses[i].ExpenseDate.Before(catExpenses[j].ExpenseDate)
		})

		// Tính khoảng cách giữa các chi tiêu
		var intervals []int
		for i := 1; i < len(catExpenses); i++ {
			interval := int(catExpenses[i].ExpenseDate.Sub(catExpenses[i-1].ExpenseDate).Hours() / 24)
			intervals = append(intervals, interval)
		}

		// Tính trung bình và độ lệch chuẩn của khoảng cách
		var sum int
		for _, interval := range intervals {
			sum += interval
		}
		avgInterval := float64(sum) / float64(len(intervals))

		var variance float64
		for _, interval := range intervals {
			variance += math.Pow(float64(interval)-avgInterval, 2)
		}
		variance /= float64(len(intervals))
		stdDev := math.Sqrt(variance)

		// Nếu độ lệch chuẩn nhỏ, có thể là chi tiêu định kỳ
		if stdDev < avgInterval*0.5 && avgInterval > 0 {
			// Tính trung bình số tiền
			var totalAmount int
			for _, expense := range catExpenses {
				totalAmount += expense.Amount
			}
			avgAmount := totalAmount / len(catExpenses)

			// Xác định tần suất
			var frequency string
			if avgInterval <= 7 {
				frequency = "weekly"
			} else if avgInterval <= 15 {
				frequency = "biweekly"
			} else if avgInterval <= 35 {
				frequency = "monthly"
			} else {
				frequency = "irregular"
			}

			recurringExpenses = append(recurringExpenses, gin.H{
				"category_id":   catID,
				"category_name": catExpenses[0].Category.Name,
				"frequency":     frequency,
				"avg_interval":  avgInterval,
				"avg_amount":    avgAmount,
				"total_amount":  totalAmount,
				"expense_count": len(catExpenses),
				"description":   "Chi tiêu định kỳ " + frequencyToVietnamese(frequency) + " cho " + catExpenses[0].Category.Name,
			})
		}
	}

	return recurringExpenses
}

// Hàm phân tích chi tiêu theo mùa
func analyzeSeasonalExpenses(expenses []models.Expense) []gin.H {
	// Nhóm chi tiêu theo tháng và danh mục
	monthCategoryExpenses := make(map[string]map[uint]int)
	for _, expense := range expenses {
		monthKey := expense.ExpenseDate.Format("2006-01")
		if _, exists := monthCategoryExpenses[monthKey]; !exists {
			monthCategoryExpenses[monthKey] = make(map[uint]int)
		}
		monthCategoryExpenses[monthKey][expense.CategoryID] += expense.Amount
	}

	// Tính trung bình chi tiêu theo danh mục
	categoryAvgExpenses := make(map[uint]int)
	categoryCount := make(map[uint]int)
	for _, categoryExpenses := range monthCategoryExpenses {
		for catID, amount := range categoryExpenses {
			categoryAvgExpenses[catID] += amount
			categoryCount[catID]++
		}
	}

	for catID := range categoryAvgExpenses {
		if categoryCount[catID] > 0 {
			categoryAvgExpenses[catID] /= categoryCount[catID]
		}
	}

	// Tìm các tháng có chi tiêu cao bất thường
	var seasonalExpenses []gin.H
	categoryNames := make(map[uint]string)
	for _, expense := range expenses {
		categoryNames[expense.CategoryID] = expense.Category.Name
	}

	for monthKey, categoryExpenses := range monthCategoryExpenses {
		for catID, amount := range categoryExpenses {
			if categoryCount[catID] < 2 {
				continue // Bỏ qua nếu không đủ dữ liệu
			}

			avgAmount := categoryAvgExpenses[catID]
			if amount > avgAmount*2 && avgAmount > 0 {
				// Chi tiêu cao gấp đôi trung bình
				month, _ := time.Parse("2006-01", monthKey)
				seasonalExpenses = append(seasonalExpenses, gin.H{
					"category_id":      catID,
					"category_name":    categoryNames[catID],
					"month":            month.Month().String(),
					"year":             month.Year(),
					"amount":           amount,
					"avg_amount":       avgAmount,
					"percent_increase": (float64(amount) - float64(avgAmount)) / float64(avgAmount) * 100,
					"description":      "Chi tiêu cao bất thường cho " + categoryNames[catID] + " trong tháng " + month.Format("01/2006"),
				})
			}
		}
	}

	return seasonalExpenses
}

// Hàm phân tích chi tiêu bốc đồng
func analyzeImpulseExpenses(expenses []models.Expense) []gin.H {
	var impulseExpenses []gin.H

	// Tìm các chi tiêu đơn lẻ có giá trị cao
	for _, expense := range expenses {
		// Lấy chi tiêu trung bình của danh mục này
		var avgAmount float64
		database.DB.Table("expenses").
			Where("category_id = ? AND id != ? AND deleted_at IS NULL", expense.CategoryID, expense.ID).
			Select("COALESCE(AVG(amount), 0)").
			Scan(&avgAmount)

		// Nếu chi tiêu cao hơn 3 lần trung bình và trung bình > 0
		if avgAmount > 0 && float64(expense.Amount) > avgAmount*3 {
			impulseExpenses = append(impulseExpenses, gin.H{
				"expense_id":     expense.ID,
				"category_id":    expense.CategoryID,
				"category_name":  expense.Category.Name,
				"amount":         expense.Amount,
				"avg_amount":     avgAmount,
				"date":           expense.ExpenseDate,
				"percent_higher": (float64(expense.Amount) - avgAmount) / avgAmount * 100,
				"description":    "Chi tiêu bốc đồng cao bất thường cho " + expense.Category.Name,
			})
		}
	}

	return impulseExpenses
}

// Hàm lưu các mẫu chi tiêu vào database
func saveExpensePatterns(userID uint, patterns []gin.H, patternType string) {
	for _, pattern := range patterns {
		categoryID := uint(pattern["category_id"].(uint))
		description := pattern["description"].(string)

		// Kiểm tra xem mẫu đã tồn tại chưa
		var existingPattern models.ExpensePattern
		result := database.DB.Where("user_id = ? AND pattern_type = ? AND category_ids LIKE ?",
			userID, patternType, "%"+strconv.FormatUint(uint64(categoryID), 10)+"%").First(&existingPattern)

		if result.RowsAffected == 0 {
			// Tạo mẫu mới
			var avgAmount int
			var frequency string

			if patternType == "recurring" {
				avgAmount = pattern["avg_amount"].(int)
				frequency = pattern["frequency"].(string)
			} else if patternType == "seasonal" {
				avgAmount = pattern["amount"].(int)
				frequency = "seasonal"
			} else {
				avgAmount = pattern["amount"].(int)
				frequency = "one-time"
			}

			newPattern := models.ExpensePattern{
				UserID:        userID,
				PatternType:   patternType,
				Description:   description,
				CategoryIDs:   strconv.FormatUint(uint64(categoryID), 10),
				AverageAmount: avgAmount,
				Frequency:     frequency,
			}

			database.DB.Create(&newPattern)
		}
	}
}

// GetAIRecommendations trả về các đề xuất tiết kiệm từ AI
func GetAIRecommendations(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Lấy các đề xuất đã lưu
	var savedRecommendations []models.AIRecommendation
	database.DB.Where("user_id = ?", userID).Find(&savedRecommendations)

	// Tạo đề xuất mới dựa trên phân tích chi tiêu
	generateNewRecommendations(userID.(uint))

	// Lấy lại tất cả đề xuất sau khi đã tạo mới
	database.DB.Where("user_id = ?", userID).Find(&savedRecommendations)

	// Thêm thông tin danh mục
	var result []gin.H
	for _, rec := range savedRecommendations {
		var category models.Category
		database.DB.First(&category, rec.CategoryID)

		result = append(result, gin.H{
			"id":               rec.ID,
			"title":            rec.Title,
			"description":      rec.Description,
			"category_id":      rec.CategoryID,
			"category_name":    category.Name,
			"potential_saving": rec.PotentialSaving,
			"implemented":      rec.Implemented,
			"created_at":       rec.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": result,
	})
}

// Hàm tạo đề xuất mới dựa trên phân tích chi tiêu
func generateNewRecommendations(userID uint) {
	// Lấy dữ liệu chi tiêu trong 3 tháng gần nhất
	now := time.Now()
	threeMonthsAgo := now.AddDate(0, -3, 0)

	// Phân tích chi tiêu theo danh mục
	type CategoryExpense struct {
		CategoryID  uint
		Name        string
		TotalAmount int
		Count       int
	}

	var categoryExpenses []CategoryExpense
	database.DB.Table("expenses").
		Select("expenses.category_id, categories.name, SUM(expenses.amount) as total_amount, COUNT(expenses.id) as count").
		Joins("JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ? AND expenses.expense_date >= ? AND expenses.deleted_at IS NULL",
			userID, threeMonthsAgo).
		Group("expenses.category_id, categories.name").
		Scan(&categoryExpenses)

	// Sắp xếp theo tổng chi tiêu giảm dần
	sort.Slice(categoryExpenses, func(i, j int) bool {
		return categoryExpenses[i].TotalAmount > categoryExpenses[j].TotalAmount
	})

	// Tạo đề xuất cho các danh mục chi tiêu cao nhất
	for i, catExpense := range categoryExpenses {
		if i >= 3 {
			break // Chỉ tạo đề xuất cho 3 danh mục chi tiêu cao nhất
		}

		// Kiểm tra xem đã có đề xuất cho danh mục này chưa
		var existingRec models.AIRecommendation
		result := database.DB.Where("user_id = ? AND category_id = ? AND created_at >= ?",
			userID, catExpense.CategoryID, now.AddDate(0, -1, 0)).First(&existingRec)

		if result.RowsAffected == 0 {
			// Tính số tiền tiết kiệm tiềm năng (10-20% tổng chi tiêu)
			savingPercent := 15.0 // 15%
			potentialSaving := int(float64(catExpense.TotalAmount) * savingPercent / 100)

			// Tạo đề xuất mới
			title := "Giảm chi tiêu cho " + catExpense.Name
			description := generateRecommendationDescription(catExpense.Name, catExpense.TotalAmount, potentialSaving)

			recommendation := models.AIRecommendation{
				UserID:          userID,
				CategoryID:      catExpense.CategoryID,
				Title:           title,
				Description:     description,
				PotentialSaving: potentialSaving,
				Implemented:     false,
			}

			database.DB.Create(&recommendation)
		}
	}

	// Tạo đề xuất dựa trên mẫu chi tiêu
	var patterns []models.ExpensePattern
	database.DB.Where("user_id = ?", userID).Find(&patterns)

	for _, pattern := range patterns {
		if pattern.PatternType == "impulse" || pattern.PatternType == "seasonal" {
			// Kiểm tra xem đã có đề xuất cho mẫu này chưa
			var existingRec models.AIRecommendation
			result := database.DB.Where("user_id = ? AND description LIKE ? AND created_at >= ?",
				userID, "%"+pattern.Description+"%", now.AddDate(0, -1, 0)).First(&existingRec)

			if result.RowsAffected == 0 {
				// Lấy ID danh mục
				categoryID, _ := strconv.ParseUint(pattern.CategoryIDs, 10, 32)

				// Lấy tên danh mục
				var category models.Category
				database.DB.First(&category, categoryID)

				// Tạo đề xuất mới
				title := ""
				description := ""
				potentialSaving := pattern.AverageAmount / 2

				if pattern.PatternType == "impulse" {
					title = "Hạn chế chi tiêu bốc đồng cho " + category.Name
					description = "Bạn có xu hướng chi tiêu bốc đồng cho " + category.Name + ". Hãy cân nhắc kỹ trước khi chi tiêu lớn cho danh mục này."
				} else {
					title = "Lập kế hoạch cho chi tiêu theo mùa"
					description = "Bạn có xu hướng chi tiêu cao cho " + category.Name + " theo mùa. Hãy lập kế hoạch tiết kiệm trước để đối phó với chi tiêu này."
				}

				recommendation := models.AIRecommendation{
					UserID:          userID,
					CategoryID:      uint(categoryID),
					Title:           title,
					Description:     description,
					PotentialSaving: potentialSaving,
					Implemented:     false,
				}

				database.DB.Create(&recommendation)
			}
		}
	}
}

// Hàm tạo mô tả đề xuất
func generateRecommendationDescription(categoryName string, totalAmount, potentialSaving int) string {
	descriptions := []string{
		"Bạn đã chi %s VND cho %s trong 3 tháng qua. Hãy cố gắng giảm chi tiêu này để tiết kiệm khoảng %s VND mỗi tháng.",
		"Chi tiêu cho %s chiếm phần lớn ngân sách của bạn (%s VND). Giảm 15%% có thể giúp bạn tiết kiệm %s VND.",
		"Hãy xem xét lại chi tiêu cho %s. Bạn đã chi %s VND và có thể tiết kiệm %s VND bằng cách tối ưu hóa chi tiêu này.",
	}

	// Chọn ngẫu nhiên một mô tả
	descIndex := time.Now().UnixNano() % int64(len(descriptions))

	// Format số tiền
	formattedTotal := formatMoney(totalAmount)
	formattedSaving := formatMoney(potentialSaving)

	return fmt.Sprintf(descriptions[descIndex], formattedTotal, categoryName, formattedSaving)
}

// Hàm chuyển đổi tần suất sang tiếng Việt
func frequencyToVietnamese(frequency string) string {
	switch frequency {
	case "daily":
		return "hàng ngày"
	case "weekly":
		return "hàng tuần"
	case "biweekly":
		return "hai tuần một lần"
	case "monthly":
		return "hàng tháng"
	default:
		return "không đều"
	}
}

// Hàm format số tiền
func formatMoney(amount int) string {
	return strconv.Itoa(amount)
}
