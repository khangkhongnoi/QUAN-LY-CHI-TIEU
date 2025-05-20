package ai

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"encoding/json"
	"math"
	"time"
)

// UserBehaviorAnalyzer phân tích hành vi chi tiêu của người dùng
type UserBehaviorAnalyzer struct {
	UserID uint
}

// UserProfile là cấu trúc hồ sơ người dùng
type UserProfile struct {
	SpendingStyle      string                 `json:"spending_style"`      // "tiết kiệm", "cân bằng", "tiêu xài"
	RiskTolerance      string                 `json:"risk_tolerance"`      // "thấp", "trung bình", "cao"
	FinancialGoals     []string               `json:"financial_goals"`     // Các mục tiêu tài chính
	SpendingCategories []CategoryPreference   `json:"spending_categories"` // Danh mục chi tiêu ưa thích
	SpendingPatterns   []SpendingPattern      `json:"spending_patterns"`   // Mẫu chi tiêu
	IncomeStability    string                 `json:"income_stability"`    // "không ổn định", "tương đối ổn định", "rất ổn định"
	SavingCapacity     int                    `json:"saving_capacity"`     // Khả năng tiết kiệm hàng tháng (VND)
	BudgetAdherence    float64                `json:"budget_adherence"`    // Mức độ tuân thủ ngân sách (0-100%)
	ImplementedTips    []ImplementedSavingTip `json:"implemented_tips"`    // Các mẹo tiết kiệm đã thực hiện
}

// CategoryPreference là cấu trúc sở thích chi tiêu theo danh mục
type CategoryPreference struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Importance   string  `json:"importance"` // "thấp", "trung bình", "cao"
	Flexibility  string  `json:"flexibility"` // "cứng", "linh hoạt", "rất linh hoạt"
	Percentage   float64 `json:"percentage"`  // Phần trăm chi tiêu
}

// SpendingPattern là cấu trúc mẫu chi tiêu
type SpendingPattern struct {
	Type        string `json:"type"`         // "định kỳ", "theo mùa", "bốc đồng"
	Description string `json:"description"`
	Frequency   string `json:"frequency,omitempty"` // Chỉ áp dụng cho chi tiêu định kỳ
}

// ImplementedSavingTip là cấu trúc mẹo tiết kiệm đã thực hiện
type ImplementedSavingTip struct {
	Title           string    `json:"title"`
	ImplementedDate time.Time `json:"implemented_date"`
	EffectiveLevel  string    `json:"effective_level"` // "thấp", "trung bình", "cao"
}

// MarketTrends là cấu trúc xu hướng thị trường
type MarketTrends struct {
	InflationRate     float64            `json:"inflation_rate"`      // Tỷ lệ lạm phát hiện tại
	InterestRates     float64            `json:"interest_rates"`      // Lãi suất ngân hàng
	ConsumerTrends    []ConsumerTrend    `json:"consumer_trends"`     // Xu hướng tiêu dùng
	SeasonalFactors   []SeasonalFactor   `json:"seasonal_factors"`    // Yếu tố theo mùa
	EconomicIndicators []EconomicIndicator `json:"economic_indicators"` // Chỉ số kinh tế
}

// ConsumerTrend là cấu trúc xu hướng tiêu dùng
type ConsumerTrend struct {
	Category    string `json:"category"`
	Trend       string `json:"trend"` // "tăng", "giảm", "ổn định"
	Description string `json:"description"`
}

// SeasonalFactor là cấu trúc yếu tố theo mùa
type SeasonalFactor struct {
	Name        string `json:"name"`
	StartMonth  int    `json:"start_month"`
	EndMonth    int    `json:"end_month"`
	Impact      string `json:"impact"` // "thấp", "trung bình", "cao"
	Description string `json:"description"`
}

// EconomicIndicator là cấu trúc chỉ số kinh tế
type EconomicIndicator struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Trend string `json:"trend"` // "tăng", "giảm", "ổn định"
}

// NewUserBehaviorAnalyzer tạo một phân tích hành vi người dùng mới
func NewUserBehaviorAnalyzer(userID uint) *UserBehaviorAnalyzer {
	return &UserBehaviorAnalyzer{
		UserID: userID,
	}
}

// GetUserProfile phân tích và tạo hồ sơ người dùng
func (a *UserBehaviorAnalyzer) GetUserProfile() (*UserProfile, error) {
	// Lấy dữ liệu trong 6 tháng gần nhất
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)

	// Tạo cấu trúc kết quả
	profile := &UserProfile{}

	// 1. Phân tích phong cách chi tiêu
	var totalExpense, totalIncome int
	database.DB.Table("expenses").
		Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL",
			a.UserID, sixMonthsAgo).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	database.DB.Table("incomes").
		Where("user_id = ? AND income_date >= ? AND deleted_at IS NULL",
			a.UserID, sixMonthsAgo).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome)

	// Tính tỷ lệ chi tiêu/thu nhập
	expenseIncomeRatio := 0.0
	if totalIncome > 0 {
		expenseIncomeRatio = float64(totalExpense) / float64(totalIncome)
	}

	// Xác định phong cách chi tiêu
	if expenseIncomeRatio < 0.6 {
		profile.SpendingStyle = "tiết kiệm"
	} else if expenseIncomeRatio < 0.85 {
		profile.SpendingStyle = "cân bằng"
	} else {
		profile.SpendingStyle = "tiêu xài"
	}

	// 2. Phân tích khả năng chấp nhận rủi ro
	var unusualExpenseCount int64
	database.DB.Model(&models.Expense{}).
		Where("user_id = ? AND expense_date >= ? AND amount > ? AND deleted_at IS NULL",
			a.UserID, sixMonthsAgo, 1000000). // Chi tiêu lớn hơn 1 triệu
		Count(&unusualExpenseCount)

	// Tính tỷ lệ chi tiêu bất thường
	unusualExpenseRatio := 0.0
	var totalExpenseCount int64
	database.DB.Model(&models.Expense{}).
		Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL",
			a.UserID, sixMonthsAgo).
		Count(&totalExpenseCount)

	if totalExpenseCount > 0 {
		unusualExpenseRatio = float64(unusualExpenseCount) / float64(totalExpenseCount)
	}

	// Xác định khả năng chấp nhận rủi ro
	if unusualExpenseRatio < 0.05 {
		profile.RiskTolerance = "thấp"
	} else if unusualExpenseRatio < 0.15 {
		profile.RiskTolerance = "trung bình"
	} else {
		profile.RiskTolerance = "cao"
	}

	// 3. Xác định mục tiêu tài chính
	var savingGoals []models.SavingGoal
	database.DB.Where("user_id = ?", a.UserID).Find(&savingGoals)

	for _, goal := range savingGoals {
		profile.FinancialGoals = append(profile.FinancialGoals, goal.Name)
	}

	// 4. Phân tích sở thích chi tiêu theo danh mục
	var categoryExpenses []struct {
		CategoryID   uint
		CategoryName string
		TotalAmount  int
		Count        int
	}

	database.DB.Table("expenses").
		Select("expenses.category_id, categories.name as category_name, SUM(expenses.amount) as total_amount, COUNT(expenses.id) as count").
		Joins("JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ? AND expenses.expense_date >= ? AND expenses.deleted_at IS NULL",
			a.UserID, sixMonthsAgo).
		Group("expenses.category_id, categories.name").
		Scan(&categoryExpenses)

	// Tính phần trăm chi tiêu theo danh mục
	for _, ce := range categoryExpenses {
		percentage := 0.0
		if totalExpense > 0 {
			percentage = float64(ce.TotalAmount) / float64(totalExpense) * 100
		}

		// Xác định mức độ quan trọng
		importance := "trung bình"
		if percentage > 30 {
			importance = "cao"
		} else if percentage < 10 {
			importance = "thấp"
		}

		// Xác định mức độ linh hoạt
		flexibility := "linh hoạt"
		
		// Kiểm tra xem có ngân sách cho danh mục này không
		var budget models.Budget
		result := database.DB.Where("user_id = ? AND category_id = ?", a.UserID, ce.CategoryID).First(&budget)
		
		if result.RowsAffected > 0 {
			// Nếu có ngân sách, kiểm tra mức độ tuân thủ
			var monthlyExpenses []struct {
				Month int
				Year  int
				Total int
			}
			
			database.DB.Table("expenses").
				Select("EXTRACT(MONTH FROM expense_date) as month, EXTRACT(YEAR FROM expense_date) as year, SUM(amount) as total").
				Where("user_id = ? AND category_id = ? AND expense_date >= ? AND deleted_at IS NULL",
					a.UserID, ce.CategoryID, sixMonthsAgo).
				Group("EXTRACT(MONTH FROM expense_date), EXTRACT(YEAR FROM expense_date)").
				Scan(&monthlyExpenses)
			
			// Tính độ lệch chuẩn của chi tiêu hàng tháng
			var monthlyAmounts []int
			for _, me := range monthlyExpenses {
				monthlyAmounts = append(monthlyAmounts, me.Total)
			}
			
			stdDev := calculateStdDev(monthlyAmounts)
			avgAmount := calculateAverage(monthlyAmounts)
			
			// Tính hệ số biến thiên (CV)
			cv := 0.0
			if avgAmount > 0 {
				cv = stdDev / avgAmount
			}
			
			// Xác định mức độ linh hoạt dựa trên hệ số biến thiên
			if cv < 0.2 {
				flexibility = "cứng" // Chi tiêu ổn định hàng tháng
			} else if cv > 0.5 {
				flexibility = "rất linh hoạt" // Chi tiêu biến động lớn
			}
		}

		profile.SpendingCategories = append(profile.SpendingCategories, CategoryPreference{
			CategoryID:   ce.CategoryID,
			CategoryName: ce.CategoryName,
			Importance:   importance,
			Flexibility:  flexibility,
			Percentage:   percentage,
		})
	}

	// 5. Phân tích mẫu chi tiêu
	var expensePatterns []models.ExpensePattern
	database.DB.Where("user_id = ?", a.UserID).Find(&expensePatterns)

	for _, pattern := range expensePatterns {
		spendingPattern := SpendingPattern{
			Description: pattern.Description,
		}

		switch pattern.PatternType {
		case "recurring":
			spendingPattern.Type = "định kỳ"
			spendingPattern.Frequency = translateFrequency(pattern.Frequency)
		case "seasonal":
			spendingPattern.Type = "theo mùa"
		case "impulse":
			spendingPattern.Type = "bốc đồng"
		}

		profile.SpendingPatterns = append(profile.SpendingPatterns, spendingPattern)
	}

	// 6. Phân tích độ ổn định thu nhập
	var monthlyIncomes []struct {
		Month int
		Year  int
		Total int
	}

	database.DB.Table("incomes").
		Select("EXTRACT(MONTH FROM income_date) as month, EXTRACT(YEAR FROM income_date) as year, SUM(amount) as total").
		Where("user_id = ? AND income_date >= ? AND deleted_at IS NULL",
			a.UserID, sixMonthsAgo).
		Group("EXTRACT(MONTH FROM income_date), EXTRACT(YEAR FROM income_date)").
		Scan(&monthlyIncomes)

	// Tính độ lệch chuẩn của thu nhập hàng tháng
	var incomeAmounts []int
	for _, mi := range monthlyIncomes {
		incomeAmounts = append(incomeAmounts, mi.Total)
	}

	incomeStdDev := calculateStdDev(incomeAmounts)
	avgIncome := calculateAverage(incomeAmounts)

	// Tính hệ số biến thiên (CV) của thu nhập
	incomeCV := 0.0
	if avgIncome > 0 {
		incomeCV = incomeStdDev / avgIncome
	}

	// Xác định độ ổn định thu nhập
	if incomeCV < 0.1 {
		profile.IncomeStability = "rất ổn định"
	} else if incomeCV < 0.3 {
		profile.IncomeStability = "tương đối ổn định"
	} else {
		profile.IncomeStability = "không ổn định"
	}

	// 7. Tính khả năng tiết kiệm hàng tháng
	avgMonthlyIncome := 0
	if len(incomeAmounts) > 0 {
		avgMonthlyIncome = int(avgIncome)
	}

	avgMonthlyExpense := 0
	if totalExpenseCount > 0 {
		// Tính số tháng
		months := int(math.Ceil(now.Sub(sixMonthsAgo).Hours() / 24 / 30))
		if months > 0 {
			avgMonthlyExpense = totalExpense / months
		}
	}

	profile.SavingCapacity = avgMonthlyIncome - avgMonthlyExpense
	if profile.SavingCapacity < 0 {
		profile.SavingCapacity = 0
	}

	// 8. Tính mức độ tuân thủ ngân sách
	var budgets []models.Budget
	database.DB.Where("user_id = ?", a.UserID).Find(&budgets)

	var totalBudgetAdherence float64
	var budgetCount int

	for _, budget := range budgets {
		var spent int
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		database.DB.Table("expenses").
			Where("category_id = ? AND user_id = ? AND expense_date >= ? AND expense_date < ? AND deleted_at IS NULL",
				budget.CategoryID, a.UserID, firstOfMonth, firstOfMonth.AddDate(0, 1, 0)).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&spent)

		// Tính mức độ tuân thủ
		adherence := 100.0
		if budget.Amount > 0 {
			adherence = 100 - math.Abs(float64(spent-budget.Amount))/float64(budget.Amount)*100
			if adherence < 0 {
				adherence = 0
			}
		}

		totalBudgetAdherence += adherence
		budgetCount++
	}

	if budgetCount > 0 {
		profile.BudgetAdherence = totalBudgetAdherence / float64(budgetCount)
	} else {
		profile.BudgetAdherence = 0
	}

	// 9. Lấy các mẹo tiết kiệm đã thực hiện
	var recommendations []models.AIRecommendation
	database.DB.Where("user_id = ? AND implemented = ?", a.UserID, true).Find(&recommendations)

	for _, rec := range recommendations {
		profile.ImplementedTips = append(profile.ImplementedTips, ImplementedSavingTip{
			Title:           rec.Title,
			ImplementedDate: rec.UpdatedAt,
			EffectiveLevel:  "trung bình", // Mặc định
		})
	}

	return profile, nil
}

// GetUserProfileJSON trả về hồ sơ người dùng dưới dạng JSON
func (a *UserBehaviorAnalyzer) GetUserProfileJSON() (string, error) {
	profile, err := a.GetUserProfile()
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GetMarketTrends lấy xu hướng thị trường hiện tại
func (a *UserBehaviorAnalyzer) GetMarketTrends() (*MarketTrends, error) {
	// Trong môi trường thực tế, dữ liệu này có thể được lấy từ API bên ngoài
	// Ở đây chúng ta sẽ sử dụng dữ liệu mẫu
	
	now := time.Now()
	currentMonth := int(now.Month())
	
	trends := &MarketTrends{
		InflationRate: 3.5,
		InterestRates: 6.0,
		ConsumerTrends: []ConsumerTrend{
			{
				Category:    "Thực phẩm",
				Trend:       "tăng",
				Description: "Giá thực phẩm đang tăng do ảnh hưởng của biến đổi khí hậu",
			},
			{
				Category:    "Công nghệ",
				Trend:       "giảm",
				Description: "Giá thiết bị điện tử đang giảm do cạnh tranh và công nghệ mới",
			},
			{
				Category:    "Giải trí",
				Trend:       "ổn định",
				Description: "Chi tiêu cho giải trí duy trì ổn định",
			},
		},
		SeasonalFactors: []SeasonalFactor{
			{
				Name:        "Tết Nguyên Đán",
				StartMonth:  1,
				EndMonth:    2,
				Impact:      "cao",
				Description: "Chi tiêu tăng mạnh cho quà tặng, thực phẩm và du lịch",
			},
			{
				Name:        "Mùa du lịch hè",
				StartMonth:  6,
				EndMonth:    8,
				Impact:      "trung bình",
				Description: "Chi tiêu tăng cho du lịch và giải trí ngoài trời",
			},
			{
				Name:        "Mùa tựu trường",
				StartMonth:  8,
				EndMonth:    9,
				Impact:      "trung bình",
				Description: "Chi tiêu tăng cho đồ dùng học tập và học phí",
			},
			{
				Name:        "Mùa mua sắm cuối năm",
				StartMonth:  11,
				EndMonth:    12,
				Impact:      "cao",
				Description: "Chi tiêu tăng cho quà tặng và mua sắm dịp lễ",
			},
		},
		EconomicIndicators: []EconomicIndicator{
			{
				Name:  "GDP",
				Value: "5.8%",
				Trend: "tăng",
			},
			{
				Name:  "Tỷ lệ thất nghiệp",
				Value: "2.5%",
				Trend: "giảm",
			},
			{
				Name:  "Tỷ giá USD/VND",
				Value: "24,500",
				Trend: "ổn định",
			},
		},
	}
	
	// Lọc các yếu tố theo mùa phù hợp với tháng hiện tại
	var relevantSeasonalFactors []SeasonalFactor
	for _, factor := range trends.SeasonalFactors {
		if (currentMonth >= factor.StartMonth && currentMonth <= factor.EndMonth) ||
			(factor.StartMonth > factor.EndMonth && (currentMonth >= factor.StartMonth || currentMonth <= factor.EndMonth)) {
			relevantSeasonalFactors = append(relevantSeasonalFactors, factor)
		}
	}
	trends.SeasonalFactors = relevantSeasonalFactors
	
	return trends, nil
}

// GetMarketTrendsJSON trả về xu hướng thị trường dưới dạng JSON
func (a *UserBehaviorAnalyzer) GetMarketTrendsJSON() (string, error) {
	trends, err := a.GetMarketTrends()
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(trends, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// Các hàm tiện ích

// calculateStdDev tính độ lệch chuẩn
func calculateStdDev(values []int) float64 {
	if len(values) == 0 {
		return 0
	}

	avg := calculateAverage(values)
	var sumSquares float64
	for _, v := range values {
		sumSquares += math.Pow(float64(v)-avg, 2)
	}

	return math.Sqrt(sumSquares / float64(len(values)))
}

// calculateAverage tính trung bình
func calculateAverage(values []int) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum int
	for _, v := range values {
		sum += v
	}

	return float64(sum) / float64(len(values))
}

// translateFrequency chuyển đổi tần suất sang tiếng Việt
func translateFrequency(frequency string) string {
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