package handlers

import (
    "database/sql"
    "DineTogether/models"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "golang.org/x/crypto/bcrypt"
    "net/http"
)

func Register(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var user models.User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }

        // 验证输入
        if user.Username == "" || user.Password == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
            return
        }

        // 加密密码
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
            return
        }

        // 默认角色为 guest，主人需手动指定
        if user.Role != "admin" {
            user.Role = "guest"
        }

        // 插入用户
        _, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
            user.Username, string(hashedPassword), user.Role)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
    }
}

func Login(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var user models.User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }

        var storedUser models.User
        row := db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", user.Username)
        if err := row.Scan(&storedUser.ID, &storedUser.Username, &storedUser.Password, &storedUser.Role); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
            return
        }

        // 验证密码
        if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
            return
        }

        // 存储 session
        session := sessions.Default(c)
        session.Set("user_id", storedUser.ID)
        session.Set("role", storedUser.Role)
        session.Save()

        c.JSON(http.StatusOK, gin.H{
            "message": "登录成功",
            "user_id": storedUser.ID,
            "role":    storedUser.Role,
        })
    }
}