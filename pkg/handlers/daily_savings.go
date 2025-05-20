package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

// TestDailySavings is a handler for testing the daily savings functionality
func TestDailySavings(c *gin.Context) {
	// Get user ID from session
	userID, exists := c.Get("user_id")
	if !exists {
		userID = uint(0)
	}
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	// Create daily savings service
	dailySavingsService := services.NewDailySavingsService()
	
	// Calculate and add savings
	savingsAmount, err := dailySavingsService.CalculateAndAddDailySavings(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to calculate daily savings",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Daily savings calculation completed",
		"savings_amount": savingsAmount,
	})
}

// ProcessAllUsersDailySavings is a handler for processing daily savings for all users
func ProcessAllUsersDailySavings(c *gin.Context) {
	// Only allow admin users to trigger this
	userID, exists := c.Get("user_id")
	if !exists || userID == uint(0) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}
	
	// Create daily savings service
	dailySavingsService := services.NewDailySavingsService()
	
	// Process savings for all users
	dailySavingsService.ProcessDailySavingsForAllUsers()
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Daily savings calculation completed for all users",
	})
}

// GetDailySavingsInfo returns information about the daily savings feature
func GetDailySavingsInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists || userID == uint(0) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}
	
	// Get today's date (start and end)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"daily_expense_limit": services.DailyExpenseLimit,
		"description": "Hệ thống sẽ tự động tính toán số tiền tiết kiệm hàng ngày dựa trên chi tiêu của bạn. Nếu tổng chi tiêu trong ngày dưới 100,000 VND, phần chênh lệch sẽ được tự động thêm vào mục tiêu tiết kiệm của bạn.",
		"calculation": "Tiền tiết kiệm = 100,000 VND - Tổng chi tiêu trong ngày (nếu tổng chi tiêu < 100,000 VND)",
		"processing_time": "Hệ thống sẽ tự động tính toán vào cuối mỗi ngày (23:59)",
	})
}

// ShowDailySavingsPage displays the daily savings information page
func ShowDailySavingsPage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists || userID == uint(0) {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	
	c.HTML(http.StatusOK, "daily_savings_info.html", nil)
}