package handlers

import (
    "database/sql"
    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
    "net/http"
)

func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        userID := session.Get("user_id")
        if userID == nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
            c.Abort()
            return
        }

        var role string
        err := db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
            c.Abort()
            return
        }

        c.Set("role", role)
        c.Set("user_id", userID)
        c.Next()
    }
}