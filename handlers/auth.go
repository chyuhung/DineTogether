package handlers

import (
	"DineTogether/errors"
	"DineTogether/models"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if err := ValidatePassword(user.Password); err != nil {
			c.Error(err)
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Error(errors.ErrInternalServer)
			return
		}
		_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", user.Username, hashedPassword, req.Role)
		if err != nil {
			c.Error(errors.NewAppError(400, "用户名已存在"))
			return
		}
		c.JSON(200, gin.H{"message": "注册成功"})
	}
}

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
			return
		}
		var user models.User
		row := db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", loginRequest.Username)
		if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
		session := sessions.Default(c)
		session.Clear()
		session.Set("user_id", user.ID)
		session.Set("role", user.Role)
		if err := session.Save(); err != nil {
			log.Printf("保存 session 失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 session 失败"})
			return
		}
		log.Printf("用户 %s 登录成功，role: %s", user.Username, user.Role)
		c.JSON(http.StatusOK, gin.H{"message": "登录成功", "user_id": user.ID, "role": user.Role})
	}
}

func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		if role != "admin" {
			log.Printf("权限验证失败，当前 role: %v", role)
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
