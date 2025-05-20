#!/bin/bash
echo "Updating email reminder feature files..."

# Replace the expense reminder service with the fixed version
rm -f pkg/services/expense_reminder_service_updated.go
mv -f pkg/services/expense_reminder_service.go.new pkg/services/expense_reminder_service.go
echo "Updated expense_reminder_service.go"

# Delete the duplicate handler file
rm -f pkg/handlers/reminder_updated.go

# Delete the duplicate main file
rm -f main_with_email_reminder.go
rm -f main_updated_with_reminder.go

echo "Email reminder feature update completed!"