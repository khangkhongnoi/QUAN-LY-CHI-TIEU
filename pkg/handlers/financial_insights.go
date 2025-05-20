package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/ai"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetFinancialInsights trả về phân tích tài chính chi tiết
func GetFinancialInsights(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo phân tích chi tiêu
	analyzer := ai.NewExpenseAnalyzer(userID.(uint))
	
	// Lấy điểm sức khỏe tài chính
	healthScore, err := analyzer.GetFinancialHealthScore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tính điểm sức khỏe tài chính"})
		return
	}
	
	// Lấy tóm tắt dữ liệu chi tiêu
	summary, err := analyzer.GetExpenseDataSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy dữ liệu chi tiêu"})
		return
	}
	
	// Tính tỷ lệ chi tiêu/thu nhập
	expenseIncomeRatio := 0.0
	if summary.TotalIncome > 0 {
		expenseIncomeRatio = float64(summary.TotalExpense) / float64(summary.TotalIncome) * 100
	}
	
	// Tạo báo cáo chi tiêu
	report, err := analyzer.GenerateExpenseReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo báo cáo chi tiêu"})
		return
	}
	
	// Trả về kết quả
	c.JSON(http.StatusOK, gin.H{
		"health_score": healthScore,
		"expense_income_ratio": expenseIncomeRatio,
		"total_expense": summary.TotalExpense,
		"total_income": summary.TotalIncome,
		"top_categories": summary.ExpenseByCategory,
		"monthly_expenses": summary.MonthlyExpenses,
		"recurring_expenses": summary.RecurringExpenses,
		"unusual_expenses": summary.UnusualExpenses,
		"budget_status": summary.Budgets,
		"saving_goals": summary.SavingGoals,
		"report": report,
	})
}

// GetFinancialHealthScore trả về điểm sức khỏe tài chính
func GetFinancialHealthScore(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo phân tích chi tiêu
	analyzer := ai.NewExpenseAnalyzer(userID.(uint))
	
	// Lấy điểm sức khỏe tài chính
	healthScore, err := analyzer.GetFinancialHealthScore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tính điểm sức khỏe tài chính"})
		return
	}
	
	// Lấy tóm tắt dữ liệu chi tiêu
	summary, err := analyzer.GetExpenseDataSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy dữ liệu chi tiêu"})
		return
	}
	
	// Tính tỷ lệ chi tiêu/thu nhập
	expenseIncomeRatio := 0.0
	if summary.TotalIncome > 0 {
		expenseIncomeRatio = float64(summary.TotalExpense) / float64(summary.TotalIncome) * 100
	}
	
	// Xác định trạng thái sức khỏe tài chính
	status := "tốt"
	if healthScore < 50 {
		status = "kém"
	} else if healthScore < 70 {
		status = "trung bình"
	}
	
	// Trả về kết quả
	c.JSON(http.StatusOK, gin.H{
		"health_score": healthScore,
		"status": status,
		"expense_income_ratio": expenseIncomeRatio,
		"total_expense": summary.TotalExpense,
		"total_income": summary.TotalIncome,
	})
}

// GetSmartBudgetSuggestions trả về đề xuất ngân sách thông minh
func GetSmartBudgetSuggestions(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Tạo adapter cho SmartSavingAdvisor
	advisor := ai.NewSmartSavingAdvisor(userID.(uint))
	
	// Tạo đề xuất tối ưu hóa ngân sách
	err := advisor.GenerateSmartBudgetOptimization()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo đề xuất ngân sách"})
		return
	}
	
	// Lấy các đề xuất đã lưu
	analyzer := ai.NewExpenseAnalyzer(userID.(uint))
	summary, err := analyzer.GetExpenseDataSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy dữ liệu chi tiêu"})
		return
	}
	
	// Trả về kết quả
	c.JSON(http.StatusOK, gin.H{
		"current_budgets": summary.Budgets,
		"expense_history": summary.ExpenseByCategory,
		"message": "Đã tạo đề xuất ngân sách thông minh. Vui lòng xem trong phần Đề xuất tiết kiệm thông minh.",
	})
}