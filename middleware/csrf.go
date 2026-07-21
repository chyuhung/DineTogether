package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GenerateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		ct := c.GetHeader("Content-Type")
		if ct != "application/json" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无效的请求 Content-Type", "success": false})
			c.Abort()
			return
		}
		session := sessions.Default(c)
		token := session.Get("csrf_token")
		if token == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token 缺失", "success": false})
			c.Abort()
			return
		}
		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken == "" || headerToken != token.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token 无效", "success": false})
			c.Abort()
			return
		}
		c.Next()
	}
}
