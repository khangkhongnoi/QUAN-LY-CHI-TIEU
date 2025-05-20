@echo off
echo Updating email reminder feature files...

rem Rename the updated expense reminder service
copy /Y pkg\services\expense_reminder_service_updated.go pkg\services\expense_reminder_service.go
echo Updated expense_reminder_service.go

rem Rename the updated reminder handler
copy /Y pkg\handlers\reminder_updated.go pkg\handlers\reminder.go
echo Updated reminder.go

rem Rename the updated main file
copy /Y main_with_email_reminder.go main.go
echo Updated main.go

echo Email reminder feature update completed!
pause