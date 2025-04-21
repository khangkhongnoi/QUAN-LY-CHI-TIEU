package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string    `gorm:"unique;not null"`
	Password    string    `gorm:"not null"`
	Email       string    `gorm:"unique"`
	Expenses    []Expense `gorm:"foreignKey:UserID"`
}

// HashPassword mã hóa mật khẩu người dùng
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword kiểm tra mật khẩu
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}