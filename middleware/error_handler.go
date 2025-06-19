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
			// 确保响应头为 application/json
			c.Header("Content-Type", "application/json")
			err := c.Errors.Last()
			if appErr, ok := err.Err.(*errors.AppError); ok {
				// 返回标准化的错误响应
				c.JSON(appErr.Code, gin.H{
					"error":   appErr.Message,
					"success": false,
				})
			} else {
				log.Printf("未处理的错误: %v", err)
				// 返回通用服务器错误
				c.JSON(errors.ErrInternalServer.Code, gin.H{
					"error":   errors.ErrInternalServer.Message,
					"success": false,
				})
			}
			// 清空错误，防止重复处理
			c.Errors = nil
			// 终止后续处理
			c.Abort()
		}
	}
}
