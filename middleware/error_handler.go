package middleware

import (
	"DineTogether/errors"
	"log"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 处理全局错误
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			if appErr, ok := err.Err.(*errors.AppError); ok {
				c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			} else {
				log.Printf("未处理的错误: %v", err)
				c.JSON(errors.ErrInternalServer.Code, gin.H{"error": "服务器内部错误"})
			}
		}
	}
}
