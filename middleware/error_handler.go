package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 处理全局错误
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			// 确保响应头为 application/json
			c.Header("Content-Type", "application/json")
			err := c.Errors.Last()
			log.Printf("未处理的错误: %v", err)
			// 返回通用服务器错误
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "服务器错误",
				"success": false,
			})
			// 清空错误，防止重复处理
			c.Errors = nil
			// 终止后续处理
			c.Abort()
		}
	}
}
