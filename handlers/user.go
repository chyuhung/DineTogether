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

func GetUserInfo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Error(errors.ErrUnauthorized)
			return
		}
		var user models.User
		row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", userID)
		if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			log.Printf("获取用户信息失败: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": user.ID, "username": user.Username, "role": user.Role})
	}
}

func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if user.Username == "" || user.Password == "" || user.Role == "" {
			c.Error(errors.NewAppError(400, "用户名、密码和角色不能为空"))
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
			user.Username, string(hashedPassword), user.Role)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: users.username" {
				c.Error(errors.NewAppError(400, "用户名已存在"))
			} else {
				log.Printf("创建用户失败: %v", err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusOK, gin.H{"message": "用户创建成功", "user_id": id})
	}
}

func GetUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, username, role FROM users")
		if err != nil {
			log.Printf("获取用户列表失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		defer rows.Close()
		var users []models.User
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
				log.Printf("扫描用户失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			users = append(users, user)
		}
		c.JSON(http.StatusOK, users)
	}
}

func GetUserByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id)
		if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			log.Printf("用户不存在: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if user.Username == "" || user.Role == "" {
			c.Error(errors.NewAppError(400, "用户名和角色不能为空"))
			return
		}
		var hashedPassword string
		if user.Password != "" {
			hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("密码加密失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			hashedPassword = string(hashedPasswordBytes)
		} else {
			row := db.QueryRow("SELECT password FROM users WHERE id = ?", id)
			if err := row.Scan(&hashedPassword); err != nil {
				log.Printf("用户不存在: %v", err)
				c.Error(errors.ErrNotFound)
				return
			}
		}
		result, err := db.Exec("UPDATE users SET username = ?, password = ?, role = ? WHERE id = ?",
			user.Username, hashedPassword, user.Role, id)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: users.username" {
				c.Error(errors.NewAppError(400, "用户名已存在"))
			} else {
				log.Printf("更新用户失败: %v", err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.Error(errors.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "用户更新成功"})
	}
}

func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			log.Printf("删除用户失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.Error(errors.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
	}
}

func ChangePassword(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Error(errors.ErrUnauthorized)
			return
		}
		var request struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if request.OldPassword == "" || request.NewPassword == "" {
			c.Error(errors.NewAppError(400, "旧密码和新密码不能为空"))
			return
		}
		if len(request.NewPassword) < 6 {
			c.Error(errors.NewAppError(400, "新密码至少6位"))
			return
		}
		var currentPassword string
		row := db.QueryRow("SELECT password FROM users WHERE id = ?", userID)
		if err := row.Scan(&currentPassword); err != nil {
			log.Printf("用户不存在: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(request.OldPassword)); err != nil {
			c.Error(errors.ErrUnauthorized)
			return
		}
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		result, err := db.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedNewPassword), userID)
		if err != nil {
			log.Printf("更新密码失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.Error(errors.ErrNotFound)
			return
		}
		log.Printf("用户 %v 修改密码成功", userID)
		c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
	}
}
