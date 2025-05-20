#!/bin/bash
echo "Updating email reminder feature files..."

# Make sure any duplicate files are removed
rm -f pkg/services/expense_reminder_service_updated.go
rm -f pkg/handlers/reminder_updated.go
rm -f main_with_email_reminder.go
rm -f main_updated_with_reminder.go

# Update main.go to include the test endpoint if it doesn't already have it
if ! grep -q "admin/test-reminder" main.go; then
  sed -i 's|authorized.GET("/financial/smart-budget", handlers.GetSmartBudgetSuggestions)|authorized.GET("/financial/smart-budget", handlers.GetSmartBudgetSuggestions)\n\t\t\t// API nhắc nhở chi tiêu (chỉ dành cho admin và testing)\n\t\t\tauthorized.GET("/admin/test-reminder", handlers.TestExpenseReminder)|' main.go
  echo "Added test reminder endpoint to main.go"
fi

echo "Email reminder feature update completed!"