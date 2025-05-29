package handlers

import (
    "database/sql"
    "DineTogether/models"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "strconv"
)

func CreateUser(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var user models.User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if user.Username == "" || user.Password == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
            return
        }
        if user.Role != "admin" && user.Role != "guest" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
            return
        }
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
            return
        }
        result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
            user.Username, string(hashedPassword), user.Role)
        if err != nil {
            if err.Error() == "UNIQUE constraint failed: users.username" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在，请选择其他名称"})
            } else {
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
            c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户失败"})
            return
        }
        defer rows.Close()
        var users []models.User
        for rows.Next() {
            var user models.User
            if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
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
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
            return
        }
        var user models.User
        row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id)
        if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
            return
        }
        c.JSON(http.StatusOK, user)
    }
}

func UpdateUser(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
            return
        }
        var user models.User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if user.Username == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
            return
        }
        if user.Role != "admin" && user.Role != "guest" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
            return
        }
        var hashedPassword string
        if user.Password != "" {
            hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
                return
            }
            hashedPassword = string(hashedPasswordBytes)
        } else {
            row := db.QueryRow("SELECT password FROM users WHERE id = ?", id)
            if err := row.Scan(&hashedPassword); err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
                return
            }
        }
        result, err := db.Exec("UPDATE users SET username = ?, password = ?, role = ? WHERE id = ?",
            user.Username, hashedPassword, user.Role, id)
        if err != nil {
            if err.Error() == "UNIQUE constraint failed: users.username" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在，请选择其他名称"})
            } else {
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
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
            return
        }
        result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
        if err != nil {
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