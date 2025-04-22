package middleware

import (
	"QUAN-LY-CHI-TIEU/pkg/database"
	"QUAN-LY-CHI-TIEU/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Tạo cache đơn giản để lưu thông tin người dùng
var (
	userCache     = make(map[uint]models.User)
	userCacheMux  = &sync.RWMutex{}
	cacheDuration = 5 * time.Minute // Thời gian cache hợp lệ
	cacheExpiry   = make(map[uint]time.Time)
)

// getUserFromCache lấy thông tin người dùng từ cache
func getUserFromCache(userID uint) (models.User, bool) {
	userCacheMux.RLock()
	defer userCacheMux.RUnlock()

	user, exists := userCache[userID]
	if !exists {
		return models.User{}, false
	}

	// Kiểm tra xem cache có hết hạn không
	expiry, ok := cacheExpiry[userID]
	if !ok || time.Now().After(expiry) {
		delete(userCache, userID)
		delete(cacheExpiry, userID)
		return models.User{}, false
	}

	return user, true
}

// setUserCache lưu thông tin người dùng vào cache
func setUserCache(user models.User) {
	userCacheMux.Lock()
	defer userCacheMux.Unlock()

	userCache[user.ID] = user
	cacheExpiry[user.ID] = time.Now().Add(cacheDuration)
}

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

		uid := uint(id)

		// Thử lấy thông tin người dùng từ cache
		user, found := getUserFromCache(uid)

		// Nếu không tìm thấy trong cache, truy vấn từ database
		if !found {
			// Tối ưu truy vấn bằng cách chỉ lấy các trường cần thiết
			result := database.DB.Select("id, username").First(&user, uid)
			if result.Error != nil {
				c.SetCookie("user_id", "", -1, "/", "", false, true)
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
				return
			}

			// Lưu vào cache để sử dụng cho các request tiếp theo
			setUserCache(user)
		}

		// Lưu thông tin user vào context
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Next()
	}
}
