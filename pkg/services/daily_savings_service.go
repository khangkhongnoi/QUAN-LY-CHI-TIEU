package services

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"log"
	"time"
)

const (
	// DailyExpenseLimit is the maximum amount allowed for daily expenses (100,000 VND)
	DailyExpenseLimit = 100000
)

// DailySavingsService handles automatic savings based on daily expenses
type DailySavingsService struct{}

// NewDailySavingsService creates a new daily savings service
func NewDailySavingsService() *DailySavingsService {
	return &DailySavingsService{}
}

// CalculateAndAddDailySavings calculates and adds savings for a user based on their daily expenses
func (s *DailySavingsService) CalculateAndAddDailySavings(userID uint) (int, error) {
	// Get today's date (start and end)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	// Calculate total expenses for today
	var totalExpenses int
	result := database.DB.Model(&models.Expense{}).
		Where("user_id = ? AND expense_date BETWEEN ? AND ?", userID, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&totalExpenses)
	
	if result.Error != nil {
		log.Printf("Error calculating total expenses: %v", result.Error)
		return 0, result.Error
	}

	// Calculate savings amount
	savingsAmount := 0
	if totalExpenses < DailyExpenseLimit {
		savingsAmount = DailyExpenseLimit - totalExpenses
	}

	// If there's something to save
	if savingsAmount > 0 {
		// Find or create a savings goal for automatic daily savings
		var savingGoal models.SavingGoal
		result := database.DB.Where("user_id = ? AND name = ?", userID, "Tiết kiệm tự động hàng ngày").First(&savingGoal)
		
		// If saving goal doesn't exist, create it
		if result.Error != nil {
			savingGoal = models.SavingGoal{
				UserID:      userID,
				Name:        "Tiết kiệm tự động hàng ngày",
				TargetAmount: 10000000, // 10 million VND as default target
				CurrentAmount: 0,
				Deadline:    time.Now().AddDate(1, 0, 0), // 1 year from now
				Description: "Tiết kiệm tự động từ chi tiêu hàng ngày dưới 100,000 VND",
			}
			
			if err := database.DB.Create(&savingGoal).Error; err != nil {
				log.Printf("Error creating saving goal: %v", err)
				return 0, err
			}
		}

		// Add a saving transaction
		savingTransaction := models.SavingTransaction{
			UserID:       userID,
			GoalID:       savingGoal.ID,
			Amount:       savingsAmount,
			Note:         "Tiết kiệm tự động: Chi tiêu hôm nay dưới 100,000 VND",
			TransactionDate: now,
		}

		if err := database.DB.Create(&savingTransaction).Error; err != nil {
			log.Printf("Error creating saving transaction: %v", err)
			return 0, err
		}

		// Update the current amount in the saving goal
		savingGoal.CurrentAmount += savingsAmount
		if err := database.DB.Save(&savingGoal).Error; err != nil {
			log.Printf("Error updating saving goal: %v", err)
			return 0, err
		}

		log.Printf("Added %d VND to daily savings for user %d", savingsAmount, userID)
	} else {
		log.Printf("No savings added for user %d. Total expenses (%d) exceeded or met the limit (%d)", 
			userID, totalExpenses, DailyExpenseLimit)
	}

	return savingsAmount, nil
}

// ProcessDailySavingsForAllUsers processes daily savings for all users
func (s *DailySavingsService) ProcessDailySavingsForAllUsers() {
	log.Println("Processing daily savings for all users")
	
	// Get all active users
	var users []*models.User
	result := database.DB.Find(&users)
	if result.Error != nil {
		log.Printf("Error fetching users: %v", result.Error)
		return
	}
	
	// Process savings for each user
	for _, user := range users {
		savingsAmount, err := s.CalculateAndAddDailySavings(user.ID)
		if err != nil {
			log.Printf("Error processing savings for user %d: %v", user.ID, err)
			continue
		}
		
		if savingsAmount > 0 {
			log.Printf("Successfully added %d VND to daily savings for user %s", savingsAmount, user.Username)
		}
	}
}