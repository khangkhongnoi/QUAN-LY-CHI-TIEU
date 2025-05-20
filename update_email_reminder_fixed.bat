@echo off
echo Updating email reminder feature files...

rem Replace the expense reminder service with the fixed version
del pkg\services\expense_reminder_service_updated.go
move /Y pkg\services\expense_reminder_service.go.new pkg\services\expense_reminder_service.go
echo Updated expense_reminder_service.go

rem Delete the duplicate handler file
del pkg\handlers\reminder_updated.go

rem Delete the duplicate main file
del main_with_email_reminder.go
del main_updated_with_reminder.go

echo Email reminder feature update completed!
pause