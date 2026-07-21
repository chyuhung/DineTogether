package handlers

import (
	"DineTogether/middleware"
	"DineTogether/models"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("密码长度必须至少6位")
	}
	return nil
}

func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if user.Username == "" || user.Password == "" {
			badRequest(c, "用户名和密码不能为空")
			return
		}
		if err := ValidatePassword(user.Password); err != nil {
			badRequest(c, err.Error())
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, 'guest')", user.Username, hashedPassword)
		if err != nil {
			if isUniqueConstraint(err) {
				badRequest(c, "用户名已存在")
			} else {
				log.Printf("注册用户失败: %v", err)
				serverError(c, "服务器错误")
			}
			return
		}
		id, _ := result.LastInsertId()
		success(c, "注册成功", gin.H{"user_id": id})
	}
}

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if loginRequest.Username == "" || loginRequest.Password == "" {
			badRequest(c, "用户名和密码不能为空")
			return
		}
		var user models.User
		row := db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", loginRequest.Username)
		if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role); err != nil {
			log.Printf("用户 %s 不存在: %v", loginRequest.Username, err)
			unauthorized(c, "用户名或密码错误")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
			log.Printf("用户 %s 密码错误", loginRequest.Username)
			unauthorized(c, "用户名或密码错误")
			return
		}
		session := sessions.Default(c)
		session.Clear()
		session.Set("user_id", user.ID)
		session.Set("role", user.Role)
		if err := session.Save(); err != nil {
			log.Printf("保存 session 失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		log.Printf("用户 %s 登录成功，角色: %s", user.Username, user.Role)
		success(c, "登录成功", gin.H{"user_id": user.ID, "role": user.Role})
	}
}

func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		userID := session.Get("user_id")
		if role != "admin" {
			forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		var exists bool
		row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID)
		if err := row.Scan(&exists); err != nil || !exists {
			session.Clear()
			session.Save()
			unauthorized(c, "用户不存在或会话已过期")
			c.Abort()
			return
		}
		c.Next()
	}
}

func Logout(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		success(c, "退出成功")
	}
}

func GetCSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("csrf_token")
		if token == nil {
			newToken := middleware.GenerateCSRFToken()
			session.Set("csrf_token", newToken)
			session.Save()
			c.JSON(http.StatusOK, gin.H{"csrf_token": newToken, "success": true})
			return
		}
		c.JSON(http.StatusOK, gin.H{"csrf_token": token.(string), "success": true})
	}
}
