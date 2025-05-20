# PowerShell script to update email reminder feature

Write-Host "Updating email reminder feature files..." -ForegroundColor Green

# Make sure any duplicate files are removed
Remove-Item -Path "pkg/services/expense_reminder_service_updated.go" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "pkg/handlers/reminder_updated.go" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "main_with_email_reminder.go" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "main_updated_with_reminder.go" -Force -ErrorAction SilentlyContinue

# Update main.go to include the test endpoint if it doesn't already have it
$mainContent = Get-Content -Path "main.go" -Raw
if (-not ($mainContent -match "admin/test-reminder")) {
    $mainContent = $mainContent -replace "authorized.GET\(`"\/financial\/smart-budget`", handlers.GetSmartBudgetSuggestions\)", "authorized.GET(`"/financial/smart-budget`", handlers.GetSmartBudgetSuggestions)`n`t`t`t// API nhắc nhở chi tiêu (chỉ dành cho admin và testing)`n`t`t`tauthorized.GET(`"/admin/test-reminder`", handlers.TestExpenseReminder)"
    Set-Content -Path "main.go" -Value $mainContent
    Write-Host "Added test reminder endpoint to main.go" -ForegroundColor Green
}

Write-Host "Email reminder feature update completed!" -ForegroundColor Green
Write-Host "Press any key to continue..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")