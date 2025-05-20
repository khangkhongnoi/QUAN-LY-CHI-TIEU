package services

import (
	"log"
	"time"
)

// SchedulerService handles scheduling and running tasks at specific times
type SchedulerService struct {
	reminderService    *ExpenseReminderService
	dailySavingsService *DailySavingsService
	stopChan           chan struct{}
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		reminderService:    NewExpenseReminderService(),
		dailySavingsService: NewDailySavingsService(),
		stopChan:           make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start() {
	log.Println("Starting expense reminder scheduler")
	
	go func() {
		for {
			select {
			case <-s.stopChan:
				log.Println("Scheduler stopped")
				return
			default:
				now := time.Now()
				hour, min, _ := now.Clock()
				
				// Check if it's one of our scheduled times for reminders
				if (hour == 21 && min == 0) || // 21:00
					(hour == 22 && min == 0) || // 22:00
					(hour == 22 && min == 30) || // 22:30
					(hour == 23 && min == 0) { // 23:00
					
					// Determine which attempt this is
					var attemptNumber int
					switch {
					case hour == 21:
						attemptNumber = 1
					case hour == 22 && min == 0:
						attemptNumber = 2
					case hour == 22 && min == 30:
						attemptNumber = 3
					case hour == 23:
						attemptNumber = 4
					}
					
					// Run the reminder check
					s.reminderService.CheckAndRemindUsers(attemptNumber)
					
					// Sleep for 1 minute to avoid running the same check multiple times
					time.Sleep(1 * time.Minute)
				} else if hour == 23 && min == 59 {
					// Process daily savings at the end of the day (23:59)
					log.Println("Running daily savings calculation for all users")
					s.dailySavingsService.ProcessDailySavingsForAllUsers()
					
					// Sleep for 1 minute to avoid running the same check multiple times
					time.Sleep(1 * time.Minute)
				} else {
					// Check every minute
					time.Sleep(1 * time.Minute)
				}
			}
		}
	}()
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	log.Println("Stopping expense reminder scheduler")
	close(s.stopChan)
}