package handlers

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// ShowLoginPage hiển thị trang đăng nhập
func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Đăng nhập",
	})
}

// ShowRegisterPage hiển thị trang đăng ký
func ShowRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title": "Đăng ký",
	})
}

// Login xử lý đăng nhập
func Login(c *gin.Context) {
	var request struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Vui lòng điền đầy đủ thông tin",
		})
		return
	}

	// Tìm user
	var user models.User
	result := database.DB.Where("username = ?", request.Username).First(&user)
	if result.Error != nil || !user.CheckPassword(request.Password) {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "Tên đăng nhập hoặc mật khẩu không đúng",
		})
		return
	}

	// Tạo cookie
	c.SetCookie("user_id", strconv.Itoa(int(user.ID)), 3600*24*30, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

// Register xử lý đăng ký
func Register(c *gin.Context) {
	var request struct {
		Username        string `form:"username" binding:"required"`
		Password        string `form:"password" binding:"required"`
		ConfirmPassword string `form:"confirm_password" binding:"required"`
		Email           string `form:"email"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"error": "Vui lòng điền đầy đủ thông tin",
		})
		return
	}

	// Kiểm tra mật khẩu xác nhận
	if request.Password != request.ConfirmPassword {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"error": "Mật khẩu xác nhận không khớp",
		})
		return
	}

	// Kiểm tra username đã tồn tại
	var existingUser models.User
	result := database.DB.Where("username = ?", request.Username).First(&existingUser)
	if result.RowsAffected > 0 {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"error": "Tên đăng nhập đã tồn tại",
		})
		return
	}

	// Tạo user mới
	user := models.User{
		Username: request.Username,
		Password: request.Password,
		Email:    request.Email,
	}

	// Hash mật khẩu
	if err := user.HashPassword(); err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"error": "Lỗi hệ thống, vui lòng thử lại sau",
		})
		return
	}

	// Lưu vào database
	result = database.DB.Create(&user)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"error": "Lỗi hệ thống, vui lòng thử lại sau",
		})
		return
	}

	// Tạo cookie và chuyển hướng
	c.SetCookie("user_id", strconv.Itoa(int(user.ID)), 3600*24*30, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

// Logout đăng xuất
func Logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}