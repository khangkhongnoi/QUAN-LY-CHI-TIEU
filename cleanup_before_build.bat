@echo off
echo Cleaning up duplicate files before building...

rem Remove any duplicate main files
del /F /Q main_updated.go 2>nul
del /F /Q main_with_email_reminder.go 2>nul
del /F /Q main_updated_with_reminder.go 2>nul
del /F /Q main_with_test_endpoint.go 2>nul

rem Remove any duplicate service files
del /F /Q pkg\services\expense_reminder_service_updated.go 2>nul
del /F /Q pkg\services\expense_reminder_service.go.new 2>nul

rem Remove any duplicate handler files
del /F /Q pkg\handlers\reminder_updated.go 2>nul

echo Cleanup completed!
echo You can now build the application with: docker-compose build
echo Or run it directly with: go run main.go
pause