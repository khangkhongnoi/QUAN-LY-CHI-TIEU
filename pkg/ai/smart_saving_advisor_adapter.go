package ai

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SmartSavingAdvisorAdapter là adapter để sử dụng LLMClient
// như một SmartSavingAdvisor để tránh phải thay đổi code trong handlers
type SmartSavingAdvisorAdapter struct {
	UserID    uint
	LLMClient *LLMClient
}

// NewSmartSavingAdvisor tạo một adapter cho LLMClient
func NewSmartSavingAdvisor(userID uint) *SmartSavingAdvisorAdapter {
	return &SmartSavingAdvisorAdapter{
		UserID:    userID,
		LLMClient: NewLLMClient(),
	}
}

// GenerateSmartSavingRecommendations tạo đề xuất tiết kiệm thông minh
func (a *SmartSavingAdvisorAdapter) GenerateSmartSavingRecommendations() error {
	// Phân tích dữ liệu chi tiêu
	analyzer := NewExpenseAnalyzer(a.UserID)
	expenseData, err := analyzer.GetExpenseDataJSON()
	if err != nil {
		return err
	}

	// Gọi API AI để lấy đề xuất
	aiResponse, err := a.LLMClient.GetSavingRecommendation(expenseData)
	if err != nil {
		return err
	}

	// Phân tích phản hồi từ AI và lưu vào database
	return a.processSavingRecommendations(aiResponse)
}

// GenerateSmartBudgetOptimization tạo đề xuất tối ưu hóa ngân sách
func (a *SmartSavingAdvisorAdapter) GenerateSmartBudgetOptimization() error {
	// Phân tích dữ liệu chi tiêu và ngân sách
	analyzer := NewExpenseAnalyzer(a.UserID)
	budgetData, err := analyzer.GetBudgetOptimizationData()
	if err != nil {
		return err
	}
	
	expenseData, err := analyzer.GetExpenseDataJSON()
	if err != nil {
		return err
	}

	// Gọi API AI để lấy đề xuất
	aiResponse, err := a.LLMClient.GetBudgetOptimization(budgetData, expenseData)
	if err != nil {
		return err
	}

	// Phân tích phản hồi từ AI và lưu vào database
	return a.processBudgetOptimization(aiResponse)
}

// GenerateComprehensiveFinancialInsights tạo phân tích tài chính toàn diện
func (a *SmartSavingAdvisorAdapter) GenerateComprehensiveFinancialInsights() error {
	// Phân tích dữ liệu tài chính
	analyzer := NewExpenseAnalyzer(a.UserID)
	financialData, err := analyzer.GetFinancialInsightsData()
	if err != nil {
		return err
	}

	// Gọi API AI để lấy phân tích
	aiResponse, err := a.LLMClient.GetFinancialInsights(financialData)
	if err != nil {
		return err
	}

	// Phân tích phản hồi từ AI và lưu vào database
	return a.processFinancialInsights(aiResponse)
}

// GetFinancialHealthScoreDetails trả về chi tiết điểm sức khỏe tài chính
func (a *SmartSavingAdvisorAdapter) GetFinancialHealthScoreDetails() (map[string]interface{}, error) {
	// Tạo một analyzer để lấy dữ liệu
	analyzer := NewExpenseAnalyzer(a.UserID)
	
	// Lấy điểm sức khỏe tài chính
	healthScore, err := analyzer.GetFinancialHealthScore()
	if err != nil {
		return nil, err
	}
	
	// Lấy tóm tắt dữ liệu chi tiêu
	summary, err := analyzer.GetExpenseDataSummary()
	if err != nil {
		return nil, err
	}
	
	// Tạo chi tiết điểm sức khỏe tài chính
	details := map[string]interface{}{
		"health_score":     healthScore,
		"total_expense":    summary.TotalExpense,
		"total_income":     summary.TotalIncome,
		"budget_count":     len(summary.Budgets),
		"saving_goal_count": len(summary.SavingGoals),
	}
	
	// Xác định trạng thái sức khỏe tài chính
	status := "tốt"
	if healthScore < 50 {
		status = "kém"
	} else if healthScore < 70 {
		status = "trung bình"
	}
	details["status"] = status
	
	return details, nil
}

// processSavingRecommendations xử lý phản hồi từ AI và lưu đề xuất tiết kiệm vào database
func (a *SmartSavingAdvisorAdapter) processSavingRecommendations(aiResponse string) error {
	// Phân tích phản hồi từ AI
	var recommendations struct {
		Recommendations []struct {
			Title        string `json:"title"`
			Description  string `json:"description"`
			SavingAmount int    `json:"saving_amount"`
			Difficulty   string `json:"difficulty"`
		} `json:"recommendations"`
	}

	err := json.Unmarshal([]byte(aiResponse), &recommendations)
	if err != nil {
		// Thử tìm và trích xuất phần JSON từ phản hồi
		jsonStr := extractJSONFromResponse(aiResponse)
		if jsonStr != "" {
			err = json.Unmarshal([]byte(jsonStr), &recommendations)
			if err != nil {
				return fmt.Errorf("không thể phân tích phản hồi từ AI: %v", err)
			}
		} else {
			return fmt.Errorf("không thể phân tích phản hồi từ AI: %v", err)
		}
	}

	// Xóa các đề xuất cũ để tránh trùng lặp
	database.DB.Where("user_id = ? AND created_at < ?", a.UserID, time.Now().AddDate(0, 0, -7)).Delete(&models.AIRecommendation{})

	// Lấy tất cả các danh mục
	var categories []models.Category
	database.DB.Find(&categories)
	
	// Tạo map để theo dõi danh mục đã được sử dụng
	usedCategories := make(map[uint]bool)
	
	// Lưu đề xuất vào database
	for _, rec := range recommendations.Recommendations {
		// Tìm danh mục phù hợp nhất dựa trên tiêu đề và nội dung
		categoryID := findBestCategoryMatch(rec.Title+" "+rec.Description, categories, usedCategories)
		
		// Tạo đề xuất mới
		recommendation := models.AIRecommendation{
			UserID:          a.UserID,
			CategoryID:      categoryID,
			Title:           rec.Title,
			Description:     rec.Description,
			PotentialSaving: rec.SavingAmount,
			Implemented:     false,
		}
		
		// Kiểm tra xem đã có đề xuất tương tự gần đây chưa
		var existingRec models.AIRecommendation
		titleLen := min(len(rec.Title), 10)
		result := database.DB.Where("user_id = ? AND title LIKE ? AND created_at >= ?",
			a.UserID, "%"+rec.Title[:titleLen]+"%", time.Now().AddDate(0, 0, -7)).First(&existingRec)
		
		if result.RowsAffected == 0 {
			database.DB.Create(&recommendation)
			// Đánh dấu danh mục đã được sử dụng
			usedCategories[categoryID] = true
		}
	}

	return nil
}

// processBudgetOptimization xử lý phản hồi từ AI và lưu đề xuất tối ưu hóa ngân sách vào database
func (a *SmartSavingAdvisorAdapter) processBudgetOptimization(aiResponse string) error {
	// Phân tích phản hồi từ AI
	var optimization struct {
		BudgetRecommendations []struct {
			Category          string `json:"category"`
			CurrentBudget     int    `json:"current_budget"`
			RecommendedBudget int    `json:"recommended_budget"`
			Reason            string `json:"reason"`
		} `json:"budget_recommendations"`
		SavingGoals []struct {
			Title       string `json:"title"`
			Amount      int    `json:"amount"`
			Duration    string `json:"duration"`
			Description string `json:"description"`
		} `json:"saving_goals"`
	}

	err := json.Unmarshal([]byte(aiResponse), &optimization)
	if err != nil {
		// Thử tìm và trích xuất phần JSON từ phản hồi
		jsonStr := extractJSONFromResponse(aiResponse)
		if jsonStr != "" {
			err = json.Unmarshal([]byte(jsonStr), &optimization)
			if err != nil {
				return fmt.Errorf("không thể phân tích phản hồi từ AI: %v", err)
			}
		} else {
			return fmt.Errorf("không thể phân tích phản hồi từ AI: %v", err)
		}
	}

	// Xóa các đề xuất ngân sách cũ để tránh trùng lặp
	database.DB.Where("user_id = ? AND created_at < ? AND title LIKE ?", 
		a.UserID, time.Now().AddDate(0, 0, -7), "Điều chỉnh ngân sách%").Delete(&models.AIRecommendation{})

	// Lấy tất cả các danh mục
	var allCategories []models.Category
	database.DB.Find(&allCategories)
	
	// Tạo map để theo dõi danh mục đã được sử dụng
	usedCategories := make(map[uint]bool)
	
	// Lưu đề xuất tối ưu hóa ngân sách
	for _, budgetRec := range optimization.BudgetRecommendations {
		// Tìm danh mục phù hợp
		var category models.Category
		result := database.DB.Where("name LIKE ?", "%"+budgetRec.Category+"%").First(&category)
		
		if result.RowsAffected > 0 {
			// Tạo đề xuất mới
			title := fmt.Sprintf("Điều chỉnh ngân sách cho %s", category.Name)
			description := fmt.Sprintf("Điều chỉnh ngân sách từ %d VND xuống %d VND. %s", 
				budgetRec.CurrentBudget, budgetRec.RecommendedBudget, budgetRec.Reason)
			
			potentialSaving := budgetRec.CurrentBudget - budgetRec.RecommendedBudget
			if potentialSaving < 0 {
				potentialSaving = 0
			}
			
			recommendation := models.AIRecommendation{
				UserID:          a.UserID,
				CategoryID:      category.ID,
				Title:           title,
				Description:     description,
				PotentialSaving: potentialSaving,
				Implemented:     false,
			}
			
			// Kiểm tra xem đã có đề xuất tương tự gần đây chưa
			var existingRec models.AIRecommendation
			titleLen := min(len(title), 15)
			result := database.DB.Where("user_id = ? AND title LIKE ? AND created_at >= ?",
				a.UserID, "%"+title[:titleLen]+"%", time.Now().AddDate(0, 0, -7)).First(&existingRec)
			
			if result.RowsAffected == 0 {
				database.DB.Create(&recommendation)
				// Đánh dấu danh mục đã được sử dụng
				usedCategories[category.ID] = true
			}
		}
	}

	// Xóa các thách thức tiết kiệm cũ để tránh trùng lặp
	database.DB.Where("user_id = ? AND created_at < ?", a.UserID, time.Now().AddDate(0, 0, -7)).Delete(&models.SavingChallenge{})

	// Lưu đề xuất mục tiêu tiết kiệm
	for _, goalRec := range optimization.SavingGoals {
		// Tìm danh mục phù hợp nhất dựa trên tiêu đề và nội dung
		categoryID := findBestCategoryMatch(goalRec.Title+" "+goalRec.Description, allCategories, usedCategories)
		
		// Tạo thách thức tiết kiệm mới
		endDate, _ := parseDuration(goalRec.Duration)
		
		challenge := models.SavingChallenge{
			UserID:        a.UserID,
			Title:         goalRec.Title,
			Description:   goalRec.Description,
			TargetAmount:  goalRec.Amount,
			CurrentAmount: 0,
			StartDate:     time.Now(),
			EndDate:       endDate,
			CategoryID:    categoryID,
			Status:        "active",
		}
		
		// Kiểm tra xem đã có thách thức tương tự gần đây chưa
		var existingChallenge models.SavingChallenge
		titleLen := min(len(goalRec.Title), 15)
		result := database.DB.Where("user_id = ? AND title LIKE ? AND created_at >= ?",
			a.UserID, "%"+goalRec.Title[:titleLen]+"%", time.Now().AddDate(0, 0, -7)).First(&existingChallenge)
		
		if result.RowsAffected == 0 {
			database.DB.Create(&challenge)
			// Đánh dấu danh mục đã được sử dụng
			usedCategories[categoryID] = true
		}
	}

	return nil
}

// processFinancialInsights xử lý phản hồi từ AI và lưu phân tích tài chính vào database
func (a *SmartSavingAdvisorAdapter) processFinancialInsights(aiResponse string) error {
	// Phân tích phản hồi từ AI
	var insights struct {
		FinancialHealthScore  int     `json:"financial_health_score"`
		SpendingIncomeRatio   float64 `json:"spending_income_ratio"`
		KeyInsights           []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Impact      string `json:"impact"`
		} `json:"key_insights"`
		Recommendations []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Priority    string `json:"priority"`
		} `json:"recommendations"`
	}

	err := json.Unmarshal([]byte(aiResponse), &insights)
	if err != nil {
		// Thử tìm và trích xuất phần JSON từ phản hồi
		jsonStr := extractJSONFromResponse(aiResponse)
		if jsonStr != "" {
			err = json.Unmarshal([]byte(jsonStr), &insights)
			if err != nil {
				return fmt.Errorf("không thể phân tích phản hồi từ AI: %v", err)
			}
		} else {
			return fmt.Errorf("không thể phân tích phản hồi từ AI: %v", err)
		}
	}

	// Xóa các đề xuất phân tích cũ để tránh trùng lặp
	database.DB.Where("user_id = ? AND created_at < ? AND category_id = ?", 
		a.UserID, time.Now().AddDate(0, 0, -7), 1).Delete(&models.AIRecommendation{})

	// Lấy tất cả các danh mục
	var categories []models.Category
	database.DB.Find(&categories)
	
	// Tạo map để theo dõi danh mục đã được sử dụng
	usedCategories := make(map[uint]bool)
	
	// Lưu các đề xuất từ phân tích
	for _, rec := range insights.Recommendations {
		// Tìm danh mục phù hợp nhất dựa trên tiêu đề và nội dung
		categoryID := findBestCategoryMatch(rec.Title+" "+rec.Description, categories, usedCategories)
		
		// Tạo đề xuất mới
		potentialSaving := 0
		if rec.Priority == "cao" {
			potentialSaving = 500000
		} else if rec.Priority == "trung bình" {
			potentialSaving = 300000
		} else {
			potentialSaving = 100000
		}
		
		recommendation := models.AIRecommendation{
			UserID:          a.UserID,
			CategoryID:      categoryID,
			Title:           rec.Title,
			Description:     rec.Description,
			PotentialSaving: potentialSaving,
			Implemented:     false,
		}
		
		// Kiểm tra xem đã có đề xuất tương tự gần đây chưa
		var existingRec models.AIRecommendation
		titleLen := min(len(rec.Title), 15)
		result := database.DB.Where("user_id = ? AND title LIKE ? AND created_at >= ?",
			a.UserID, "%"+rec.Title[:titleLen]+"%", time.Now().AddDate(0, 0, -7)).First(&existingRec)
		
		if result.RowsAffected == 0 {
			database.DB.Create(&recommendation)
			// Đánh dấu danh mục đã được sử dụng
			usedCategories[categoryID] = true
		}
	}

	return nil
}

// Hàm tiện ích

// extractJSONFromResponse trích xuất phần JSON từ phản hồi
func extractJSONFromResponse(response string) string {
	jsonStart := 0
	jsonEnd := len(response)
	
	// Tìm dấu { đầu tiên
	for i, c := range response {
		if c == '{' {
			jsonStart = i
			break
		}
	}
	
	// Tìm dấu } cuối cùng
	for i := len(response) - 1; i >= 0; i-- {
		if response[i] == '}' {
			jsonEnd = i + 1
			break
		}
	}
	
	if jsonStart < jsonEnd {
		return response[jsonStart:jsonEnd]
	}
	
	return ""
}

// findBestCategoryMatch tìm danh mục phù hợp nhất dựa trên nội dung và tránh trùng lặp
func findBestCategoryMatch(content string, categories []models.Category, usedCategories map[uint]bool) uint {
	// Chuyển nội dung về chữ thường để so sánh
	content = strings.ToLower(content)
	
	// Tìm danh mục phù hợp nhất dựa trên nội dung
	var bestCategoryID uint = 0
	var bestMatch float64 = 0
	
	// Danh sách các từ khóa liên quan đến các danh mục phổ biến
	categoryKeywords := map[string][]string{
		"ăn uống":      {"ăn", "uống", "nhà hàng", "quán ăn", "thức ăn", "đồ ăn", "cà phê", "cafe", "ẩm thực"},
		"mua sắm":      {"mua", "sắm", "quần áo", "giày dép", "thời trang", "shopping"},
		"giải trí":     {"giải trí", "xem phim", "du lịch", "chơi game", "phim", "nhạc", "concert", "sở thích"},
		"đi lại":       {"đi lại", "xăng", "xe", "taxi", "grab", "giao thông", "phương tiện", "đi chuyển"},
		"hóa đơn":      {"hóa đơn", "điện", "nước", "internet", "điện thoại", "gas", "tiện ích"},
		"nhà cửa":      {"nhà", "thuê nhà", "sửa chữa", "đồ gia dụng", "nội thất", "trang trí"},
		"sức khỏe":     {"sức khỏe", "thuốc", "bệnh viện", "khám bệnh", "bảo hiểm", "gym", "thể dục"},
		"giáo dục":     {"học", "sách", "khóa học", "trường", "giáo dục", "đào tạo"},
		"tiết kiệm":    {"tiết kiệm", "đầu tư", "quỹ", "tiền gửi", "chứng khoán", "tích lũy"},
		"khác":         {"khác", "chi tiêu", "chi phí", "phát sinh"},
	}
	
	// Đầu tiên, thử tìm danh mục dựa trên từ khóa
	for _, cat := range categories {
		// Bỏ qua danh mục đã được sử dụng nếu có các lựa chọn khác
		if usedCategories[cat.ID] && len(categories) > len(usedCategories) {
			continue
		}
		
		catName := strings.ToLower(cat.Name)
		
		// Kiểm tra xem danh mục có từ khóa nào trong nội dung không
		if keywords, exists := categoryKeywords[catName]; exists {
			for _, keyword := range keywords {
				if strings.Contains(content, keyword) {
					// Nếu tìm thấy từ khóa, ưu tiên danh mục này
					return cat.ID
				}
			}
		}
		
		// Nếu tên danh mục xuất hiện trong nội dung
		if strings.Contains(content, catName) {
			return cat.ID
		}
		
		// Tính độ tương đồng
		similarity := calculateSimilarity(content, catName)
		if similarity > bestMatch {
			bestMatch = similarity
			bestCategoryID = cat.ID
		}
	}
	
	// Nếu không tìm thấy danh mục phù hợp, thử tìm danh mục chưa được sử dụng
	if bestCategoryID == 0 || usedCategories[bestCategoryID] {
		for _, cat := range categories {
			if !usedCategories[cat.ID] {
				return cat.ID
			}
		}
	}
	
	// Nếu vẫn không tìm thấy, sử dụng danh mục đầu tiên
	if bestCategoryID == 0 && len(categories) > 0 {
		bestCategoryID = categories[0].ID
	}
	
	return bestCategoryID
}

// calculateSimilarity tính độ tương đồng giữa hai chuỗi
func calculateSimilarity(s1, s2 string) float64 {
	// Đơn giản hóa: tính tỷ lệ ký tự chung
	if len(s1) == 0 || len(s2) == 0 {
		return 0
	}
	
	s1Map := make(map[rune]bool)
	for _, c := range s1 {
		s1Map[c] = true
	}
	
	var common int
	for _, c := range s2 {
		if s1Map[c] {
			common++
		}
	}
	
	return float64(common) / float64(max(len(s1), len(s2)))
}

// parseDuration phân tích chuỗi thời gian thành thời điểm
func parseDuration(duration string) (time.Time, error) {
	// Mặc định là 3 tháng
	months := 3
	
	// Thử phân tích chuỗi
	fmt.Sscanf(duration, "%d tháng", &months)
	
	return time.Now().AddDate(0, months, 0), nil
}

// min trả về giá trị nhỏ hơn
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max trả về giá trị lớn hơn
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}