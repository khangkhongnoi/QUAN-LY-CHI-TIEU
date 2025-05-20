package services

import (
	"QUAN-LY-CHI-TIEU/pkg/models"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// EmailService handles sending emails to users
type EmailService struct {
	smtpServer   string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	return &EmailService{
		smtpServer:   getEnvWithDefault("SMTP_SERVER", "smtp.gmail.com"),
		smtpPort:     getEnvWithDefault("SMTP_PORT", "587"),
		smtpUsername: getEnvWithDefault("SMTP_USERNAME", "cvkhang@vttu.edu.vn"),
		smtpPassword: getEnvWithDefault("SMTP_PASSWORD", "fcdk fycz exqm duix"),
		fromEmail:    getEnvWithDefault("FROM_EMAIL", "your-email@gmail.com"),
	}
}

// SendExpenseReminderEmail sends a reminder email to a user who hasn't added expenses today
func (s *EmailService) SendExpenseReminderEmail(user *models.User, attemptNumber int) error {
	to := []string{user.Email}

	// Create email subject based on attempt number
	var subject string
	switch attemptNumber {
	case 1:
		subject = "Nhắc nhở: Bạn chưa thêm chi tiêu hôm nay"
	case 2:
		subject = "Nhắc nhở lần 2: Bạn chưa thêm chi tiêu hôm nay"
	case 3:
		subject = "Nhắc nhở lần 3: Bạn chưa thêm chi tiêu hôm nay"
	default:
		subject = "Nhắc nhở cuối: Bạn chưa thêm chi tiêu hôm nay"
	}

	// Create email body
	body := fmt.Sprintf(`
Xin chào %s,

Chúng tôi nhận thấy bạn chưa thêm bất kỳ chi tiêu nào cho ngày hôm nay.

Việc ghi chép chi tiêu đều đặn sẽ giúp bạn theo dõi tài chính hiệu quả hơn và đạt được các mục tiêu tài chính của mình.

Vui lòng đăng nhập vào ứng dụng Quản Lý Chi Tiêu để cập nhật chi tiêu của bạn. http://113.164.79.241:8202

Trân trọng,
Đội ngũ Quản Lý Chi Tiêu
`, user.Username)

	// Construct email message
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message := fmt.Sprintf("Subject: %s\n%s\n%s", subject, mime, body)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpServer)
	addr := fmt.Sprintf("%s:%s", s.smtpServer, s.smtpPort)

	// Send email
	err := smtp.SendMail(addr, auth, s.fromEmail, to, []byte(message))
	if err != nil {
		log.Printf("Error sending email to %s: %v", user.Email, err)
		return err
	}

	log.Printf("Reminder email sent to %s (attempt #%d)", user.Email, attemptNumber)
	return nil
}

// SendBulkExpenseReminderEmails sends reminder emails to multiple users
func (s *EmailService) SendBulkExpenseReminderEmails(users []*models.User, attemptNumber int) {
	for _, user := range users {
		// Skip users without email
		if user.Email == "" || !strings.Contains(user.Email, "@") {
			continue
		}

		err := s.SendExpenseReminderEmail(user, attemptNumber)
		if err != nil {
			log.Printf("Failed to send reminder email to user %s: %v", user.Username, err)
		}
	}
}

// Helper function to get environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
