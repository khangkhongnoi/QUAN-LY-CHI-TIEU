package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/ai"
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// GetSmartSavingRecommendations trả về đề xuất tiết kiệm thông minh nâng cao
func GetSmartSavingRecommendations(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo advisor tiết kiệm thông minh
	advisor := ai.NewSmartSavingAdvisor(userID.(uint))
	
	// Kiểm tra xem có cần tạo đề xuất mới không
	var savedRecommendations []models.AIRecommendation
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&savedRecommendations)
	
	if len(savedRecommendations) == 0 || isRecommendationOutdated(savedRecommendations) {
		// Tạo đề xuất tiết kiệm thông minh mới
		err := advisor.GenerateSmartSavingRecommendations()
		if err != nil {
			log.Printf("Lỗi khi tạo đề xuất tiết kiệm thông minh: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo đề xuất tiết kiệm thông minh"})
			return
		}
		
		// Lấy lại danh sách đề xuất sau khi tạo mới
		database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&savedRecommendations)
	}
	
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

// GetSmartBudgetOptimization trả về đề xuất tối ưu hóa ngân sách thông minh
func GetSmartBudgetOptimization(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo advisor tiết kiệm thông minh
	advisor := ai.NewSmartSavingAdvisor(userID.(uint))
	
	// Tạo đề xuất tối ưu hóa ngân sách
	err := advisor.GenerateSmartBudgetOptimization()
	if err != nil {
		log.Printf("Lỗi khi tạo đề xuất tối ưu hóa ngân sách: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo đề xuất tối ưu hóa ngân sách"})
		return
	}
	
	// Lấy thông tin ngân sách hiện tại
	var budgets []models.Budget
	database.DB.Where("user_id = ?", userID).Find(&budgets)
	
	var budgetSummary []gin.H
	for _, budget := range budgets {
		var category models.Category
		database.DB.First(&category, budget.CategoryID)
		
		// Tính chi tiêu hiện tại
		var spent int
		now := time.Now()
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		
		database.DB.Table("expenses").
			Where("category_id = ? AND user_id = ? AND expense_date >= ? AND expense_date < ? AND deleted_at IS NULL",
				budget.CategoryID, userID, firstOfMonth, firstOfMonth.AddDate(0, 1, 0)).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&spent)
		
		// Tính phần trăm đã chi tiêu
		percentage := 0.0
		if budget.Amount > 0 {
			percentage = float64(spent) / float64(budget.Amount) * 100
		}
		
		budgetSummary = append(budgetSummary, gin.H{
			"id":            budget.ID,
			"category_id":   budget.CategoryID,
			"category_name": category.Name,
			"amount":        budget.Amount,
			"spent":         spent,
			"percentage":    percentage,
		})
	}
	
	// Lấy thông tin thách thức tiết kiệm
	var challenges []models.SavingChallenge
	database.DB.Where("user_id = ? AND status = ?", userID, "active").Find(&challenges)
	
	var challengeSummary []gin.H
	for _, challenge := range challenges {
		var category models.Category
		database.DB.First(&category, challenge.CategoryID)
		
		// Tính phần trăm hoàn thành
		percentage := 0.0
		if challenge.TargetAmount > 0 {
			percentage = float64(challenge.CurrentAmount) / float64(challenge.TargetAmount) * 100
		}
		
		challengeSummary = append(challengeSummary, gin.H{
			"id":               challenge.ID,
			"title":            challenge.Title,
			"description":      challenge.Description,
			"target_amount":    challenge.TargetAmount,
			"current_amount":   challenge.CurrentAmount,
			"start_date":       challenge.StartDate,
			"end_date":         challenge.EndDate,
			"category_name":    category.Name,
			"percentage":       percentage,
			"days_remaining":   int(challenge.EndDate.Sub(time.Now()).Hours() / 24),
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"current_budgets":    budgetSummary,
		"saving_challenges":  challengeSummary,
		"message":            "Đã tạo đề xuất tối ưu hóa ngân sách thông minh. Vui lòng xem trong phần Đề xuất tiết kiệm thông minh.",
	})
}

// GetComprehensiveFinancialInsights trả về phân tích tài chính toàn diện
func GetComprehensiveFinancialInsights(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo advisor tiết kiệm thông minh
	advisor := ai.NewSmartSavingAdvisor(userID.(uint))
	
	// Tạo phân tích tài chính toàn diện
	err := advisor.GenerateComprehensiveFinancialInsights()
	if err != nil {
		log.Printf("Lỗi khi tạo phân tích tài chính toàn diện: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo phân tích tài chính toàn diện"})
		return
	}
	
	// Lấy chi tiết điểm sức khỏe tài chính
	healthDetails, err := advisor.GetFinancialHealthScoreDetails()
	if err != nil {
		log.Printf("Lỗi khi lấy chi tiết điểm sức khỏe tài chính: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy chi tiết điểm sức khỏe tài chính"})
		return
	}
	
	// Lấy các cảnh báo rủi ro
	var riskWarnings []models.BudgetWarning
	database.DB.Where("user_id = ? AND warning_type = ? AND dismissed = ?", userID, "risk_factor", false).Find(&riskWarnings)
	
	var risks []gin.H
	for _, warning := range riskWarnings {
		risks = append(risks, gin.H{
			"id":        warning.ID,
			"message":   warning.Message,
			"threshold": warning.Threshold,
			"severity":  getSeverityFromThreshold(warning.Threshold),
		})
	}
	
	// Lấy hồ sơ người dùng
	behaviorAnalyzer := ai.NewUserBehaviorAnalyzer(userID.(uint))
	userProfile, err := behaviorAnalyzer.GetUserProfile()
	if err != nil {
		log.Printf("Lỗi khi lấy hồ sơ người dùng: %v", err)
	}
	
	// Lấy xu hướng thị trường
	marketTrends, err := behaviorAnalyzer.GetMarketTrends()
	if err != nil {
		log.Printf("Lỗi khi lấy xu hướng thị trường: %v", err)
	}
	
	// Tạo phản hồi
	response := gin.H{
		"health_details": healthDetails,
		"risk_factors":   risks,
	}
	
	// Thêm thông tin hồ sơ người dùng nếu có
	if userProfile != nil {
		response["user_profile"] = gin.H{
			"spending_style":   userProfile.SpendingStyle,
			"risk_tolerance":   userProfile.RiskTolerance,
			"income_stability": userProfile.IncomeStability,
			"saving_capacity":  userProfile.SavingCapacity,
			"budget_adherence": userProfile.BudgetAdherence,
		}
	}
	
	// Thêm thông tin xu hướng thị trường nếu có
	if marketTrends != nil {
		response["market_trends"] = gin.H{
			"inflation_rate": marketTrends.InflationRate,
			"interest_rates": marketTrends.InterestRates,
			"seasonal_factors": marketTrends.SeasonalFactors,
		}
	}
	
	c.JSON(http.StatusOK, response)
}

// GetUserBehaviorProfile trả về hồ sơ hành vi người dùng
func GetUserBehaviorProfile(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo phân tích hành vi người dùng
	analyzer := ai.NewUserBehaviorAnalyzer(userID.(uint))
	
	// Lấy hồ sơ người dùng
	profile, err := analyzer.GetUserProfile()
	if err != nil {
		log.Printf("Lỗi khi lấy hồ sơ người dùng: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy hồ sơ người dùng"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}

// GetMarketTrendsInfo trả về thông tin xu hướng thị trường
func GetMarketTrendsInfo(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo phân tích hành vi người dùng
	analyzer := ai.NewUserBehaviorAnalyzer(userID.(uint))
	
	// Lấy xu hướng thị trường
	trends, err := analyzer.GetMarketTrends()
	if err != nil {
		log.Printf("Lỗi khi lấy xu hướng thị trường: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy xu hướng thị trường"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
	})
}

// ToggleRecommendationImplementation chuyển đổi trạng thái thực hiện đề xuất
func ToggleRecommendationImplementation(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	// Lấy ID đề xuất và trạng thái mới
	var request struct {
		ID          uint `json:"id" binding:"required"`
		Implemented bool `json:"implemented"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Tìm đề xuất
	var recommendation models.AIRecommendation
	result := database.DB.Where("id = ? AND user_id = ?", request.ID, userID).First(&recommendation)
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy đề xuất"})
		return
	}
	
	// Cập nhật trạng thái
	recommendation.Implemented = request.Implemented
	database.DB.Save(&recommendation)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật trạng thái đề xuất thành công",
		"recommendation": gin.H{
			"id":          recommendation.ID,
			"implemented": recommendation.Implemented,
		},
	})
}

// DismissRiskWarning bỏ qua cảnh báo rủi ro
func DismissRiskWarning(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	// Lấy ID cảnh báo
	var request struct {
		ID uint `json:"id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Tìm cảnh báo
	var warning models.BudgetWarning
	result := database.DB.Where("id = ? AND user_id = ?", request.ID, userID).First(&warning)
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy cảnh báo"})
		return
	}
	
	// Cập nhật trạng thái
	warning.Dismissed = true
	database.DB.Save(&warning)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Đã bỏ qua cảnh báo",
	})
}

// Hàm tiện ích

// getSeverityFromThreshold chuyển đổi ngưỡng thành mức độ nghiêm trọng
func getSeverityFromThreshold(threshold int) string {
	if threshold >= 90 {
		return "cao"
	} else if threshold >= 70 {
		return "trung bình"
	} else {
		return "thấp"
	}
}