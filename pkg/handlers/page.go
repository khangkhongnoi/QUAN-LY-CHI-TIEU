package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ShowAdvancedPage hiển thị trang tính năng nâng cao
func ShowAdvancedPage(c *gin.Context) {
	c.HTML(http.StatusOK, "advanced.html", nil)
}