package handlers

import (
	"DineTogether/errors"
	"DineTogether/models"
	"database/sql"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ValidatePassword 验证密码是否符合要求（至少6位，包含字母和数字）
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.NewAppError(http.StatusBadRequest, "密码长度必须至少6位")
	}
	hasLetter, _ := regexp.MatchString("[a-zA-Z]", password)
	hasNumber, _ := regexp.MatchString("[0-9]", password)
	if !hasLetter || !hasNumber {
		return errors.NewAppError(http.StatusBadRequest, "密码必须包含字母和数字")
	}
	return nil
}

// Register 处理用户注册
func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if user.Username == "" || user.Password == "" {
			c.Error(errors.NewAppError(http.StatusBadRequest, "用户名和密码不能为空"))
			return
		}
		if err := ValidatePassword(user.Password); err != nil {
			c.Error(err)
			return
		}
		user.Role = "guest" // 固定角色为普通用户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", user.Username, hashedPassword, user.Role)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: users.username" {
				c.Error(errors.NewAppError(http.StatusBadRequest, "用户名已存在"))
			} else {
				log.Printf("注册用户失败: %v", err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusOK, gin.H{"message": "注册成功", "user_id": id})
	}
}

// Login 处理用户登录
func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if loginRequest.Username == "" || loginRequest.Password == "" {
			c.Error(errors.NewAppError(http.StatusBadRequest, "用户名和密码不能为空"))
			return
		}
		var user models.User
		row := db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", loginRequest.Username)
		if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role); err != nil {
			log.Printf("用户 %s 不存在: %v", loginRequest.Username, err)
			c.Error(errors.ErrUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
			log.Printf("用户 %s 密码错误", loginRequest.Username)
			c.Error(errors.ErrUnauthorized)
			return
		}
		session := sessions.Default(c)
		session.Clear()
		session.Set("user_id", user.ID)
		session.Set("role", user.Role)
		if err := session.Save(); err != nil {
			log.Printf("保存 session 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		log.Printf("用户 %s 登录成功，角色: %s", user.Username, user.Role)
		c.JSON(http.StatusOK, gin.H{"message": "登录成功", "user_id": user.ID, "role": user.Role})
	}
}

// AuthMiddleware 验证管理员权限
func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		if role != "admin" {
			log.Printf("权限验证失败，当前角色: %v", role)
			c.Error(errors.NewAppError(http.StatusForbidden, "需要管理员权限"))
			c.Abort()
			return
		}
		c.Next()
	}
}
