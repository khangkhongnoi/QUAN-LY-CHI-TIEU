#!/bin/bash
echo "Updating email reminder feature files..."

# Rename the updated expense reminder service
cp -f pkg/services/expense_reminder_service_updated.go pkg/services/expense_reminder_service.go
echo "Updated expense_reminder_service.go"

# Rename the updated reminder handler
cp -f pkg/handlers/reminder_updated.go pkg/handlers/reminder.go
echo "Updated reminder.go"

# Rename the updated main file
cp -f main_with_email_reminder.go main.go
echo "Updated main.go"

echo "Email reminder feature update completed!"