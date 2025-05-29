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
        role := session.Get("role")
        if role != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
            c.Abort()
            return
        }
        c.Next()
    }
}