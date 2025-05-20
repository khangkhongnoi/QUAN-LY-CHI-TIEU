package services

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"log"
	"time"
)

// ExpenseReminderService handles checking and reminding users to add expenses
type ExpenseReminderService struct {
	emailService *EmailService
}

// NewExpenseReminderService creates a new expense reminder service
func NewExpenseReminderService() *ExpenseReminderService {
	return &ExpenseReminderService{
		emailService: NewEmailService(),
	}
}

// CheckAndRemindUsers checks if users have added expenses today and sends reminders
func (s *ExpenseReminderService) CheckAndRemindUsers(attemptNumber int) {
	log.Printf("Running expense reminder check (attempt #%d)", attemptNumber)
	
	// Get all active users with email addresses
	var users []*models.User
	result := database.DB.Find(&users)
	if result.Error != nil {
		log.Printf("Error fetching users: %v", result.Error)
		return
	}
	
	// Get today's date (start and end)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	
	// Users who need reminders
	var usersNeedingReminders []*models.User
	
	// Check each user
	for _, user := range users {
		// Skip users without email
		if user.Email == "" {
			continue
		}
		
		// Check if user has added any expenses today
		var count int64
		database.DB.Model(&models.Expense{}).
			Where("user_id = ? AND expense_date BETWEEN ? AND ?", user.ID, startOfDay, endOfDay).
			Count(&count)
		
		// If no expenses found, add user to reminder list
		if count == 0 {
			usersNeedingReminders = append(usersNeedingReminders, user)
		}
	}
	
	// Send reminders to users who haven't added expenses
	if len(usersNeedingReminders) > 0 {
		log.Printf("Sending reminders to %d users", len(usersNeedingReminders))
		s.emailService.SendBulkExpenseReminderEmails(usersNeedingReminders, attemptNumber)
	} else {
		log.Println("All users have added expenses today. No reminders needed.")
	}
}