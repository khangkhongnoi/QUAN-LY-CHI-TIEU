package middleware

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// AuthRequired kiểm tra xác thực người dùng
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy cookie
		userID, err := c.Cookie("user_id")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Chuyển đổi userID từ string sang uint
		id, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			c.SetCookie("user_id", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Kiểm tra user có tồn tại
		var user models.User
		result := database.DB.First(&user, id)
		if result.Error != nil {
			c.SetCookie("user_id", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Lưu thông tin user vào context
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Next()
	}
}