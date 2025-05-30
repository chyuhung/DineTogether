package handlers

import (
    "database/sql"
    "DineTogether/models"
    "github.com/gin-gonic/gin"
    "log"
    "net/http"
    "strconv"
)

func CreateMenu(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var menu models.Menu
        if err := c.ShouldBindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if menu.Name == "" || menu.EnergyCost <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "菜品名称和精力消耗不能为空或无效"})
            return
        }
        result, err := db.Exec("INSERT INTO menus (name, description, energy_cost) VALUES (?, ?, ?)",
            menu.Name, menu.Description, menu.EnergyCost)
        if err != nil {
            if err.Error() == "UNIQUE constraint failed: menus.name" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "菜品名称已存在"})
            } else {
                log.Printf("创建菜品失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "创建菜品失败"})
            }
            return
        }
        id, _ := result.LastInsertId()
        c.JSON(http.StatusOK, gin.H{"message": "菜品创建成功", "menu_id": id})
    }
}

func GetMenus(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        rows, err := db.Query("SELECT id, name, description, energy_cost FROM menus")
        if err != nil {
            log.Printf("获取菜品失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "获取菜品失败"})
            return
        }
        defer rows.Close()
        var menus []models.Menu
        for rows.Next() {
            var menu models.Menu
            if err := rows.Scan(&menu.ID, &menu.Name, &menu.Description, &menu.EnergyCost); err != nil {
                log.Printf("扫描菜品失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描菜品失败"})
                return
            }
            menus = append(menus, menu)
        }
        c.JSON(http.StatusOK, menus)
    }
}

func GetMenuByID(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的菜品 ID"})
            return
        }
        var menu models.Menu
        row := db.QueryRow("SELECT id, name, description, energy_cost FROM menus WHERE id = ?", id)
        if err := row.Scan(&menu.ID, &menu.Name, &menu.Description, &menu.EnergyCost); err != nil {
            log.Printf("菜品不存在: %v", err)
            c.JSON(http.StatusNotFound, gin.H{"error": "菜品不存在"})
            return
        }
        c.JSON(http.StatusOK, menu)
    }
}

func UpdateMenu(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的菜品 ID"})
            return
        }
        var menu models.Menu
        if err := c.ShouldBindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if menu.Name == "" || menu.EnergyCost <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "菜品名称和精力消耗不能为空或无效"})
            return
        }
        result, err := db.Exec("UPDATE menus SET name = ?, description = ?, energy_cost = ? WHERE id = ?",
            menu.Name, menu.Description, menu.EnergyCost, id)
        if err != nil {
            if err.Error() == "UNIQUE constraint failed: menus.name" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "菜品名称已存在"})
            } else {
                log.Printf("更新菜品失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "更新菜品失败"})
            }
            return
        }
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "菜品不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "菜品更新成功"})
    }
}

func DeleteMenu(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的菜品 ID"})
            return
        }
        result, err := db.Exec("DELETE FROM menus WHERE id = ?", id)
        if err != nil {
            log.Printf("删除菜品失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "删除菜品失败"})
            return
        }
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "菜品不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "菜品删除成功"})
    }
}