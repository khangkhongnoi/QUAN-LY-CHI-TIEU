#!/bin/bash
echo "Cleaning up duplicate files before building..."

# Remove any duplicate main files
rm -f main_updated.go
rm -f main_with_email_reminder.go
rm -f main_updated_with_reminder.go
rm -f main_with_test_endpoint.go

# Remove any duplicate service files
rm -f pkg/services/expense_reminder_service_updated.go
rm -f pkg/services/expense_reminder_service.go.new

# Remove any duplicate handler files
rm -f pkg/handlers/reminder_updated.go

echo "Cleanup completed!"
echo "You can now build the application with: docker-compose build"
echo "Or run it directly with: go run main.go"