package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success 200 OK，直接返回数据对象
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Error 返回 HTTP 错误码 + 标准 JSON 错误信息
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}
