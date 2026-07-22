package handlers

import (
	"DineTogether/models"
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUserInfo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			unauthorized(c, "未授权")
			return
		}
		var user models.User
		row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", userID)
		if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			log.Printf("获取用户 %v 信息失败: %v", userID, err)
			notFound(c, "资源未找到")
			return
		}
		success(c, "获取用户信息成功", gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		})
	}
}

func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if user.Username == "" || user.Password == "" || user.Role == "" {
			badRequest(c, "用户名、密码和角色不能为空")
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
		result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", user.Username, hashedPassword, user.Role)
		if err != nil {
			if isUniqueConstraint(err) {
				badRequest(c, "用户名已存在")
			} else {
				log.Printf("创建用户失败: %v", err)
				serverError(c, "服务器错误")
			}
			return
		}
		id, _ := result.LastInsertId()
		success(c, "用户创建成功", gin.H{"user_id": id})
	}
}

func GetUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, username, role FROM users")
		if err != nil {
			log.Printf("获取用户列表失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
				log.Printf("扫描用户失败: %v", err)
				serverError(c, "服务器错误")
				return
			}
			users = append(users, user)
		}
		if err := rows.Err(); err != nil {
			log.Printf("遍历用户行失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		success(c, "获取用户列表成功", gin.H{"users": users})
	}
}

func GetUserByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id)
		if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			log.Printf("用户 %s 不存在: %v", id, err)
			notFound(c, "资源未找到")
			return
		}
		success(c, "获取用户信息成功", gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		})
	}
}

func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if user.Username == "" || user.Role == "" {
			badRequest(c, "用户名和角色不能为空")
			return
		}
		var hashedPassword string
		if user.Password != "" {
			if err := ValidatePassword(user.Password); err != nil {
				badRequest(c, err.Error())
				return
			}
			hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("密码加密失败: %v", err)
				serverError(c, "服务器错误")
				return
			}
			hashedPassword = string(hashedPasswordBytes)
		} else {
			row := db.QueryRow("SELECT password FROM users WHERE id = ?", id)
			if err := row.Scan(&hashedPassword); err != nil {
				log.Printf("用户 %s 不存在: %v", id, err)
				notFound(c, "资源未找到")
				return
			}
		}
		result, err := db.Exec("UPDATE users SET username = ?, password = ?, role = ? WHERE id = ?", user.Username, hashedPassword, user.Role, id)
		if err != nil {
			if isUniqueConstraint(err) {
				badRequest(c, "用户名已存在")
			} else {
				log.Printf("更新用户 %s 失败: %v", id, err)
				serverError(c, "服务器错误")
			}
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "资源未找到")
			return
		}
		success(c, "用户更新成功")
	}
}

func UpdateUserRole(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Role string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || (req.Role != "admin" && req.Role != "guest") {
			badRequest(c, "无效的角色")
			return
		}
		session := sessions.Default(c)
		currentUserID := session.Get("user_id")
		if currentUserID != nil {
			currStr := fmt.Sprintf("%v", currentUserID)
			if currStr == id {
				badRequest(c, "不能修改自己的角色")
				return
			}
		}
		result, err := db.Exec("UPDATE users SET role = ? WHERE id = ?", req.Role, id)
		if err != nil {
			log.Printf("更新用户 %s 角色失败: %v", id, err)
			serverError(c, "服务器错误")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "资源未找到")
			return
		}
		label := "管理员"
		if req.Role == "guest" {
			label = "普通用户"
		}
		success(c, fmt.Sprintf("已设为%s", label))
	}
}

func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			log.Printf("删除用户 %s 失败: %v", id, err)
			serverError(c, "服务器错误")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "资源未找到")
			return
		}
		success(c, "用户删除成功")
	}
}

func ChangePassword(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			unauthorized(c, "用户未登录")
			return
		}
		var request struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if request.OldPassword == "" || request.NewPassword == "" {
			badRequest(c, "旧密码和新密码不能为空")
			return
		}
		if err := ValidatePassword(request.NewPassword); err != nil {
			badRequest(c, err.Error())
			return
		}
		var currentPassword string
		row := db.QueryRow("SELECT password FROM users WHERE id = ?", userID)
		if err := row.Scan(&currentPassword); err != nil {
			log.Printf("用户 %v 不存在: %v", userID, err)
			notFound(c, "用户不存在")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(request.OldPassword)); err != nil {
			log.Printf("用户 %v 旧密码错误", userID)
			unauthorized(c, "旧密码错误")
			return
		}
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		result, err := db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedNewPassword, userID)
		if err != nil {
			log.Printf("更新用户 %v 密码失败: %v", userID, err)
			serverError(c, "服务器错误")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "用户不存在")
			return
		}
		log.Printf("用户 %v 修改密码成功", userID)
		success(c, "密码修改成功")
	}
}
