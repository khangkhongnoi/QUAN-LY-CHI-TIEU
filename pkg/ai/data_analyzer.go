package ai

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// ExpenseAnalyzer phân tích dữ liệu chi tiêu
type ExpenseAnalyzer struct {
	UserID uint
}

// ExpenseDataSummary là cấu trúc tóm tắt dữ liệu chi tiêu
type ExpenseDataSummary struct {
	TotalExpense      int                      `json:"total_expense"`
	TotalIncome       int                      `json:"total_income"`
	ExpenseByCategory []CategoryExpenseSummary `json:"expense_by_category"`
	MonthlyExpenses   []MonthlyExpenseSummary  `json:"monthly_expenses"`
	RecurringExpenses []RecurringExpense       `json:"recurring_expenses"`
	UnusualExpenses   []UnusualExpense         `json:"unusual_expenses"`
	Budgets           []BudgetSummary          `json:"budgets"`
	SavingGoals       []SavingGoalSummary      `json:"saving_goals"`
}

// CategoryExpenseSummary là cấu trúc tóm tắt chi tiêu theo danh mục
type CategoryExpenseSummary struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  int     `json:"total_amount"`
	Percentage   float64 `json:"percentage"`
	Count        int     `json:"count"`
	AvgAmount    int     `json:"avg_amount"`
}

// MonthlyExpenseSummary là cấu trúc tóm tắt chi tiêu theo tháng
type MonthlyExpenseSummary struct {
	Month       int `json:"month"`
	Year        int `json:"year"`
	TotalAmount int `json:"total_amount"`
}

// RecurringExpense là cấu trúc chi tiêu định kỳ
type RecurringExpense struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	Frequency    string `json:"frequency"`
	AvgAmount    int    `json:"avg_amount"`
}

// UnusualExpense là cấu trúc chi tiêu bất thường
type UnusualExpense struct {
	CategoryID    uint      `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	Amount        int       `json:"amount"`
	Date          time.Time `json:"date"`
	AvgAmount     int       `json:"avg_amount"`
	PercentHigher float64   `json:"percent_higher"`
}

// BudgetSummary là cấu trúc tóm tắt ngân sách
type BudgetSummary struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       int     `json:"amount"`
	SpentAmount  int     `json:"spent_amount"`
	Percentage   float64 `json:"percentage"`
}

// SavingGoalSummary là cấu trúc tóm tắt mục tiêu tiết kiệm
type SavingGoalSummary struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	TargetAmount  int       `json:"target_amount"`
	CurrentAmount int       `json:"current_amount"`
	CreatedAt     time.Time `json:"created_at"`
	Deadline      time.Time `json:"deadline"`
	Percentage    float64   `json:"percentage"`
}

// NewExpenseAnalyzer tạo một phân tích chi tiêu mới
func NewExpenseAnalyzer(userID uint) *ExpenseAnalyzer {
	return &ExpenseAnalyzer{
		UserID: userID,
	}
}

// GetExpenseDataSummary lấy tóm tắt dữ liệu chi tiêu
func (a *ExpenseAnalyzer) GetExpenseDataSummary() (*ExpenseDataSummary, error) {
	// Lấy dữ liệu trong 3 tháng gần nhất
	now := time.Now()
	threeMonthsAgo := now.AddDate(0, -3, 0)

	// Tạo cấu trúc kết quả
	summary := &ExpenseDataSummary{}

	// 1. Lấy tổng chi tiêu
	database.DB.Table("expenses").
		Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL",
			a.UserID, threeMonthsAgo).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&summary.TotalExpense)

	// 2. Lấy tổng thu nhập
	database.DB.Table("incomes").
		Where("user_id = ? AND income_date >= ? AND deleted_at IS NULL",
			a.UserID, threeMonthsAgo).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&summary.TotalIncome)

	// 3. Lấy chi tiêu theo danh mục
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
			a.UserID, threeMonthsAgo).
		Group("expenses.category_id, categories.name").
		Scan(&categoryExpenses)

	// Tính phần trăm và trung bình
	for _, ce := range categoryExpenses {
		percentage := 0.0
		if summary.TotalExpense > 0 {
			percentage = float64(ce.TotalAmount) / float64(summary.TotalExpense) * 100
		}

		avgAmount := 0
		if ce.Count > 0 {
			avgAmount = ce.TotalAmount / ce.Count
		}

		summary.ExpenseByCategory = append(summary.ExpenseByCategory, CategoryExpenseSummary{
			CategoryID:   ce.CategoryID,
			CategoryName: ce.CategoryName,
			TotalAmount:  ce.TotalAmount,
			Percentage:   percentage,
			Count:        ce.Count,
			AvgAmount:    avgAmount,
		})
	}

	// Sắp xếp theo tổng chi tiêu giảm dần
	sort.Slice(summary.ExpenseByCategory, func(i, j int) bool {
		return summary.ExpenseByCategory[i].TotalAmount > summary.ExpenseByCategory[j].TotalAmount
	})

	// 4. Lấy chi tiêu theo tháng
	var monthlyExpenses []struct {
		Month       int
		Year        int
		TotalAmount int
	}

	database.DB.Table("expenses").
		Select("EXTRACT(MONTH FROM expense_date) as month, EXTRACT(YEAR FROM expense_date) as year, SUM(amount) as total_amount").
		Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL",
			a.UserID, threeMonthsAgo).
		Group("EXTRACT(MONTH FROM expense_date), EXTRACT(YEAR FROM expense_date)").
		Scan(&monthlyExpenses)

	for _, me := range monthlyExpenses {
		summary.MonthlyExpenses = append(summary.MonthlyExpenses, MonthlyExpenseSummary{
			Month:       me.Month,
			Year:        me.Year,
			TotalAmount: me.TotalAmount,
		})
	}

	// 5. Phân tích chi tiêu định kỳ
	recurringExpenses := a.analyzeRecurringExpenses(threeMonthsAgo)
	summary.RecurringExpenses = recurringExpenses

	// 6. Phát hiện chi tiêu bất thường
	unusualExpenses := a.detectUnusualExpenses(threeMonthsAgo)
	summary.UnusualExpenses = unusualExpenses

	// 7. Lấy thông tin ngân sách
	var budgets []models.Budget
	database.DB.Where("user_id = ?", a.UserID).Find(&budgets)

	for _, budget := range budgets {
		var spent int
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		database.DB.Table("expenses").
			Where("category_id = ? AND user_id = ? AND expense_date >= ? AND expense_date < ? AND deleted_at IS NULL",
				budget.CategoryID, a.UserID, firstOfMonth, firstOfMonth.AddDate(0, 1, 0)).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&spent)

		percentage := 0.0
		if budget.Amount > 0 {
			percentage = float64(spent) / float64(budget.Amount) * 100
		}

		var category models.Category
		database.DB.First(&category, budget.CategoryID)

		summary.Budgets = append(summary.Budgets, BudgetSummary{
			CategoryID:   budget.CategoryID,
			CategoryName: category.Name,
			Amount:       budget.Amount,
			SpentAmount:  spent,
			Percentage:   percentage,
		})
	}

	// 8. Lấy thông tin mục tiêu tiết kiệm
	var savingGoals []models.SavingGoal
	database.DB.Where("user_id = ?", a.UserID).Find(&savingGoals)

	for _, goal := range savingGoals {
		percentage := 0.0
		if goal.TargetAmount > 0 {
			percentage = float64(goal.CurrentAmount) / float64(goal.TargetAmount) * 100
		}

		summary.SavingGoals = append(summary.SavingGoals, SavingGoalSummary{
			ID:            goal.ID,
			Name:          goal.Name,
			TargetAmount:  goal.TargetAmount,
			CurrentAmount: goal.CurrentAmount,
			CreatedAt:     goal.CreatedAt,
			Deadline:      goal.Deadline,
			Percentage:    percentage,
		})
	}

	return summary, nil
}

// analyzeRecurringExpenses phân tích chi tiêu định kỳ
func (a *ExpenseAnalyzer) analyzeRecurringExpenses(startDate time.Time) []RecurringExpense {
	// Lấy tất cả chi tiêu trong khoảng thời gian
	var expenses []models.Expense
	database.DB.Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL",
		a.UserID, startDate).
		Preload("Category").Find(&expenses)

	// Nhóm chi tiêu theo danh mục
	categoryExpenses := make(map[uint][]models.Expense)
	for _, expense := range expenses {
		categoryExpenses[expense.CategoryID] = append(categoryExpenses[expense.CategoryID], expense)
	}

	var recurringExpenses []RecurringExpense

	// Phân tích từng danh mục
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

		// Xác định tần suất dựa trên khoảng cách trung bình
		frequency := "irregular"
		if stdDev < avgInterval*0.5 { // Nếu độ lệch chuẩn nhỏ, chi tiêu có tính định kỳ
			if avgInterval >= 25 && avgInterval <= 35 {
				frequency = "monthly"
			} else if avgInterval >= 6 && avgInterval <= 8 {
				frequency = "weekly"
			} else if avgInterval >= 13 && avgInterval <= 16 {
				frequency = "biweekly"
			} else if avgInterval >= 1 && avgInterval <= 3 {
				frequency = "daily"
			}
		}

		// Nếu có tính định kỳ, thêm vào kết quả
		if frequency != "irregular" {
			// Tính trung bình chi tiêu
			var totalAmount int
			for _, expense := range catExpenses {
				totalAmount += expense.Amount
			}
			avgAmount := totalAmount / len(catExpenses)

			recurringExpenses = append(recurringExpenses, RecurringExpense{
				CategoryID:   catID,
				CategoryName: catExpenses[0].Category.Name,
				Frequency:    frequency,
				AvgAmount:    avgAmount,
			})
		}
	}

	return recurringExpenses
}

// detectUnusualExpenses phát hiện chi tiêu bất thường
func (a *ExpenseAnalyzer) detectUnusualExpenses(startDate time.Time) []UnusualExpense {
	// Lấy tất cả chi tiêu trong khoảng thời gian
	var expenses []models.Expense
	database.DB.Where("user_id = ? AND expense_date >= ? AND deleted_at IS NULL",
		a.UserID, startDate).
		Preload("Category").Find(&expenses)

	// Nhóm chi tiêu theo danh mục
	categoryExpenses := make(map[uint][]models.Expense)
	for _, expense := range expenses {
		categoryExpenses[expense.CategoryID] = append(categoryExpenses[expense.CategoryID], expense)
	}

	var unusualExpenses []UnusualExpense

	// Phân tích từng danh mục
	for catID, catExpenses := range categoryExpenses {
		if len(catExpenses) < 3 {
			continue // Bỏ qua nếu không đủ dữ liệu
		}

		// Tính trung bình chi tiêu
		var totalAmount int
		for _, expense := range catExpenses {
			totalAmount += expense.Amount
		}
		avgAmount := totalAmount / len(catExpenses)

		// Tìm chi tiêu bất thường (vượt quá 150% trung bình)
		for _, expense := range catExpenses {
			if float64(expense.Amount) > float64(avgAmount)*1.5 && avgAmount > 0 {
				percentHigher := (float64(expense.Amount) - float64(avgAmount)) / float64(avgAmount) * 100

				unusualExpenses = append(unusualExpenses, UnusualExpense{
					CategoryID:    catID,
					CategoryName:  expense.Category.Name,
					Amount:        expense.Amount,
					Date:          expense.ExpenseDate,
					AvgAmount:     avgAmount,
					PercentHigher: percentHigher,
				})
			}
		}
	}

	return unusualExpenses
}

// GetExpenseDataJSON trả về dữ liệu chi tiêu dưới dạng JSON
func (a *ExpenseAnalyzer) GetExpenseDataJSON() (string, error) {
	summary, err := a.GetExpenseDataSummary()
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GetFinancialHealthScore tính điểm sức khỏe tài chính
func (a *ExpenseAnalyzer) GetFinancialHealthScore() (int, error) {
	summary, err := a.GetExpenseDataSummary()
	if err != nil {
		return 0, err
	}

	// Tính điểm dựa trên nhiều yếu tố
	score := 50 // Điểm cơ bản

	// 1. Tỷ lệ chi tiêu/thu nhập (tối đa 30 điểm)
	expenseIncomeRatio := 0.0
	if summary.TotalIncome > 0 {
		expenseIncomeRatio = float64(summary.TotalExpense) / float64(summary.TotalIncome)
	}

	if expenseIncomeRatio <= 0.5 {
		score += 30 // Rất tốt: chi tiêu ít hơn 50% thu nhập
	} else if expenseIncomeRatio <= 0.7 {
		score += 20 // Tốt: chi tiêu 50-70% thu nhập
	} else if expenseIncomeRatio <= 0.9 {
		score += 10 // Trung bình: chi tiêu 70-90% thu nhập
	} else if expenseIncomeRatio > 1 {
		score -= 10 // Kém: chi tiêu vượt quá thu nhập
	}

	// 2. Ngân sách (tối đa 10 điểm)
	budgetExceeded := false
	for _, budget := range summary.Budgets {
		if budget.Percentage > 100 {
			budgetExceeded = true
			break
		}
	}

	if !budgetExceeded && len(summary.Budgets) > 0 {
		score += 10 // Không vượt quá ngân sách
	} else if budgetExceeded {
		score -= 5 // Vượt quá ngân sách
	}

	// 3. Mục tiêu tiết kiệm (tối đa 10 điểm)
	if len(summary.SavingGoals) > 0 {
		score += 10 // Có mục tiêu tiết kiệm
	}

	// Giới hạn điểm trong khoảng 0-100
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score, nil
}

// GetFinancialInsightsData tạo dữ liệu phân tích tài chính
func (a *ExpenseAnalyzer) GetFinancialInsightsData() (string, error) {
	summary, err := a.GetExpenseDataSummary()
	if err != nil {
		return "", err
	}

	healthScore, err := a.GetFinancialHealthScore()
	if err != nil {
		return "", err
	}

	// Tạo cấu trúc dữ liệu phân tích
	type FinancialInsights struct {
		HealthScore        int                      `json:"health_score"`
		ExpenseIncomeRatio float64                  `json:"expense_income_ratio"`
		TotalExpense       int                      `json:"total_expense"`
		TotalIncome        int                      `json:"total_income"`
		TopCategories      []CategoryExpenseSummary `json:"top_categories"`
		RecurringExpenses  []RecurringExpense       `json:"recurring_expenses"`
		UnusualExpenses    []UnusualExpense         `json:"unusual_expenses"`
		BudgetStatus       []BudgetSummary          `json:"budget_status"`
		SavingGoals        []SavingGoalSummary      `json:"saving_goals"`
	}

	// Tính tỷ lệ chi tiêu/thu nhập
	expenseIncomeRatio := 0.0
	if summary.TotalIncome > 0 {
		expenseIncomeRatio = float64(summary.TotalExpense) / float64(summary.TotalIncome)
	}

	// Lấy top 5 danh mục chi tiêu
	topCategories := []CategoryExpenseSummary{}
	if len(summary.ExpenseByCategory) > 5 {
		topCategories = summary.ExpenseByCategory[:5]
	} else {
		topCategories = summary.ExpenseByCategory
	}

	insights := FinancialInsights{
		HealthScore:        healthScore,
		ExpenseIncomeRatio: expenseIncomeRatio,
		TotalExpense:       summary.TotalExpense,
		TotalIncome:        summary.TotalIncome,
		TopCategories:      topCategories,
		RecurringExpenses:  summary.RecurringExpenses,
		UnusualExpenses:    summary.UnusualExpenses,
		BudgetStatus:       summary.Budgets,
		SavingGoals:        summary.SavingGoals,
	}

	jsonData, err := json.MarshalIndent(insights, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GetBudgetOptimizationData tạo dữ liệu tối ưu hóa ngân sách
func (a *ExpenseAnalyzer) GetBudgetOptimizationData() (string, error) {
	summary, err := a.GetExpenseDataSummary()
	if err != nil {
		return "", err
	}

	// Tạo cấu trúc dữ liệu tối ưu hóa ngân sách
	type BudgetOptimizationData struct {
		CurrentBudgets  []BudgetSummary          `json:"current_budgets"`
		ExpenseHistory  []CategoryExpenseSummary `json:"expense_history"`
		MonthlyExpenses []MonthlyExpenseSummary  `json:"monthly_expenses"`
		TotalIncome     int                      `json:"total_income"`
		TotalExpense    int                      `json:"total_expense"`
	}

	data := BudgetOptimizationData{
		CurrentBudgets:  summary.Budgets,
		ExpenseHistory:  summary.ExpenseByCategory,
		MonthlyExpenses: summary.MonthlyExpenses,
		TotalIncome:     summary.TotalIncome,
		TotalExpense:    summary.TotalExpense,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GenerateExpenseReport tạo báo cáo chi tiêu
func (a *ExpenseAnalyzer) GenerateExpenseReport() (string, error) {
	summary, err := a.GetExpenseDataSummary()
	if err != nil {
		return "", err
	}

	healthScore, err := a.GetFinancialHealthScore()
	if err != nil {
		return "", err
	}

	// Tạo báo cáo dạng văn bản
	report := fmt.Sprintf(`
BÁO CÁO CHI TIÊU (3 THÁNG GẦN NHẤT)
===================================

Điểm sức khỏe tài chính: %d/100

TỔNG QUAN:
- Tổng chi tiêu: %d VND
- Tổng thu nhập: %d VND
- Tỷ lệ chi tiêu/thu nhập: %.2f%%

TOP DANH MỤC CHI TIÊU:
`, healthScore, summary.TotalExpense, summary.TotalIncome, float64(summary.TotalExpense)/float64(summary.TotalIncome)*100)

	// Thêm thông tin danh mục chi tiêu
	for i, category := range summary.ExpenseByCategory {
		if i >= 5 {
			break // Chỉ hiển thị top 5
		}
		report += fmt.Sprintf("- %s: %d VND (%.2f%%)\n", category.CategoryName, category.TotalAmount, category.Percentage)
	}

	report += "\nCHI TIÊU ĐỊNH KỲ:\n"
	if len(summary.RecurringExpenses) > 0 {
		for _, recurring := range summary.RecurringExpenses {
			report += fmt.Sprintf("- %s: %d VND (%s)\n", recurring.CategoryName, recurring.AvgAmount, recurring.Frequency)
		}
	} else {
		report += "- Không phát hiện chi tiêu định kỳ\n"
	}

	report += "\nCHI TIÊU BẤT THƯỜNG:\n"
	if len(summary.UnusualExpenses) > 0 {
		for _, unusual := range summary.UnusualExpenses {
			report += fmt.Sprintf("- %s: %d VND (cao hơn %.2f%% so với trung bình)\n",
				unusual.CategoryName, unusual.Amount, unusual.PercentHigher)
		}
	} else {
		report += "- Không phát hiện chi tiêu bất thường\n"
	}

	report += "\nTRẠNG THÁI NGÂN SÁCH:\n"
	if len(summary.Budgets) > 0 {
		for _, budget := range summary.Budgets {
			report += fmt.Sprintf("- %s: %d/%d VND (%.2f%%)\n",
				budget.CategoryName, budget.SpentAmount, budget.Amount, budget.Percentage)
		}
	} else {
		report += "- Chưa thiết lập ngân sách\n"
	}

	report += "\nMỤC TIÊU TIẾT KIỆM:\n"
	if len(summary.SavingGoals) > 0 {
		for _, goal := range summary.SavingGoals {
			report += fmt.Sprintf("- %s: %d/%d VND (%.2f%%)\n",
				goal.Name, goal.CurrentAmount, goal.TargetAmount, goal.Percentage)
		}
	} else {
		report += "- Chưa thiết lập mục tiêu tiết kiệm\n"
	}

	return report, nil
}
