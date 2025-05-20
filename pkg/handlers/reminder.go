package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// TestExpenseReminder is a handler for testing the expense reminder functionality
func TestExpenseReminder(c *gin.Context) {
	// Get attempt number from query parameter (default to 1)
	attemptStr := c.DefaultQuery("attempt", "1")
	attempt, err := strconv.Atoi(attemptStr)
	if err != nil || attempt < 1 || attempt > 4 {
		attempt = 1
	}
	
	// Create reminder service and run check
	reminderService := services.NewExpenseReminderService()
	reminderService.CheckAndRemindUsers(attempt)
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Expense reminder check triggered successfully",
		"attempt": attempt,
	})
}