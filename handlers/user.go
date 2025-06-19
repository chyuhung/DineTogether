package handlers

import (
	"DineTogether/models"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUserInfo 获取当前用户信息
func GetUserInfo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权", "success": false})
			return
		}
		var user models.User
		row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", userID)
		if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			log.Printf("获取用户 %v 信息失败: %v", userID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到", "success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "获取用户信息成功",
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		})
	}
}

// CreateUser 创建新用户
func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "success": false})
			return
		}
		if user.Username == "" || user.Password == "" || user.Role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名、密码和角色不能为空", "success": false})
			return
		}
		if err := ValidatePassword(user.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", user.Username, hashedPassword, user.Role)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: users.username" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在", "success": false})
			} else {
				log.Printf("创建用户失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			}
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusOK, gin.H{"message": "用户创建成功", "user_id": id})
	}
}

// GetUsers 获取用户列表
func GetUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, username, role FROM users")
		if err != nil {
			log.Printf("获取用户列表失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		defer rows.Close()
		var users []models.User
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
				log.Printf("扫描用户失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
				return
			}
			users = append(users, user)
		}
		c.JSON(http.StatusOK, gin.H{"message": "获取用户列表成功", "users": users})
	}
}

// GetUserByID 获取指定用户信息
func GetUserByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id)
		if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			log.Printf("用户 %s 不存在: %v", id, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到", "success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "获取用户信息成功",
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		})
	}
}

// UpdateUser 更新用户信息
func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "success": false})
			return
		}
		if user.Username == "" || user.Role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和角色不能为空", "success": false})
			return
		}
		var hashedPassword string
		if user.Password != "" {
			if err := ValidatePassword(user.Password); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
				return
			}
			hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("密码加密失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
				return
			}
			hashedPassword = string(hashedPasswordBytes)
		} else {
			row := db.QueryRow("SELECT password FROM users WHERE id = ?", id)
			if err := row.Scan(&hashedPassword); err != nil {
				log.Printf("用户 %s 不存在: %v", id, err)
				c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到", "success": false})
				return
			}
		}
		result, err := db.Exec("UPDATE users SET username = ?, password = ?, role = ? WHERE id = ?", user.Username, hashedPassword, user.Role, id)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: users.username" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在", "success": false})
			} else {
				log.Printf("更新用户 %s 失败: %v", id, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			}
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到", "success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "用户更新成功"})
	}
}

// DeleteUser 删除用户
func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			log.Printf("删除用户 %s 失败: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到", "success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
	}
}

// ChangePassword 修改用户密码
func ChangePassword(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录", "success": false})
			return
		}
		var request struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "success": false})
			return
		}
		if request.OldPassword == "" || request.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "旧密码和新密码不能为空", "success": false})
			return
		}
		if err := ValidatePassword(request.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}
		var currentPassword string
		row := db.QueryRow("SELECT password FROM users WHERE id = ?", userID)
		if err := row.Scan(&currentPassword); err != nil {
			log.Printf("用户 %v 不存在: %v", userID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在", "success": false})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(request.OldPassword)); err != nil {
			log.Printf("用户 %v 旧密码错误", userID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码错误", "success": false})
			return
		}
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		result, err := db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedNewPassword, userID)
		if err != nil {
			log.Printf("更新用户 %v 密码失败: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在", "success": false})
			return
		}
		log.Printf("用户 %v 修改密码成功", userID)
		c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
	}
}
