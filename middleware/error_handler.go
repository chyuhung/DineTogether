package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.Header("Content-Type", "application/json")
			err := c.Errors.Last()
			log.Printf("未处理的错误: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "服务器错误",
				"success": false,
			})
			c.Errors = nil
			c.Abort()
		}
	}
}
