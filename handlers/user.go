package handlers

import (
    "database/sql"
    "DineTogether/models"
    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "log"
    "net/http"
)


func GetUserInfo(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        userID := session.Get("user_id")
        if userID == nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
            return
        }
        var user models.User
        row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", userID)
        if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
            log.Printf("获取用户信息失败: %v", err)
            c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"id": user.ID, "username": user.Username, "role": user.Role})
    }
}

func CreateUser(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var user models.User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if user.Username == "" || user.Password == "" || user.Role == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名、密码和角色不能为空"})
            return
        }
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
        if err != nil {
            log.Printf("密码加密失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
            return
        }
        result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
            user.Username, string(hashedPassword), user.Role)
        if err != nil {
            if err.Error() == "UNIQUE constraint failed: users.username" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
            } else {
                log.Printf("创建用户失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
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
            c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
            return
        }
        defer rows.Close()
        var users []models.User
        for rows.Next() {
            var user models.User
            if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
                log.Printf("扫描用户失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描用户失败"})
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
            c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
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
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if user.Username == "" || user.Role == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和角色不能为空"})
            return
        }
        var hashedPassword string
        if user.Password != "" {
            hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
            if err != nil {
                log.Printf("密码加密失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
                return
            }
            hashedPassword = string(hashedPasswordBytes)
        } else {
            row := db.QueryRow("SELECT password FROM users WHERE id = ?", id)
            if err := row.Scan(&hashedPassword); err != nil {
                log.Printf("用户不存在: %v", err)
                c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
                return
            }
        }
        result, err := db.Exec("UPDATE users SET username = ?, password = ?, role = ? WHERE id = ?",
            user.Username, hashedPassword, user.Role, id)
        if err != nil {
            if err.Error() == "UNIQUE constraint failed: users.username" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
            } else {
                log.Printf("更新用户失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
            }
            return
        }
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
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
            c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
            return
        }
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
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
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
            return
        }
        var request struct {
            OldPassword string `json:"old_password"`
            NewPassword string `json:"new_password"`
        }
        if err := c.ShouldBindJSON(&request); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if request.OldPassword == "" || request.NewPassword == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "旧密码和新密码不能为空"})
            return
        }
        if len(request.NewPassword) < 6 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "新密码至少6位"})
            return
        }
        // 获取当前密码
        var currentPassword string
        row := db.QueryRow("SELECT password FROM users WHERE id = ?", userID)
        if err := row.Scan(&currentPassword); err != nil {
            log.Printf("用户不存在: %v", err)
            c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            return
        }
        // 验证旧密码
        if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(request.OldPassword)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码错误"})
            return
        }
        // 加密新密码
        hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
        if err != nil {
            log.Printf("密码加密失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
            return
        }
        // 更新密码
        result, err := db.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedNewPassword), userID)
        if err != nil {
            log.Printf("更新密码失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
            return
        }
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            return
        }
        log.Printf("用户 %v 修改密码成功", userID)
        c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
    }
}