package handlers

import (
    "database/sql"
    "DineTogether/models"
    "github.com/gin-gonic/gin"
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
            c.JSON(http.StatusBadRequest, gin.H{"error": "菜品名称和能量值不能为空或无效"})
            return
        }
        // 检查菜名是否已存在
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM menus WHERE name = ?", menu.Name).Scan(&count)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "检查菜名失败"})
            return
        }
        if count > 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "菜名已存在，请选择其他名称"})
            return
        }
        result, err := db.Exec("INSERT INTO menus (name, description, energy_cost) VALUES (?, ?, ?)",
            menu.Name, menu.Description, menu.EnergyCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "创建菜谱失败"})
            return
        }
        id, _ := result.LastInsertId()
        c.JSON(http.StatusOK, gin.H{"message": "菜谱创建成功", "menu_id": id})
    }
}

func GetMenus(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        rows, err := db.Query("SELECT id, name, description, energy_cost FROM menus")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "获取菜谱失败"})
            return
        }
        defer rows.Close()
        var menus []models.Menu
        for rows.Next() {
            var menu models.Menu
            if err := rows.Scan(&menu.ID, &menu.Name, &menu.Description, &menu.EnergyCost); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描菜谱失败"})
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
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的菜谱 ID"})
            return
        }
        var menu models.Menu
        row := db.QueryRow("SELECT id, name, description, energy_cost FROM menus WHERE id = ?", id)
        if err := row.Scan(&menu.ID, &menu.Name, &menu.Description, &menu.EnergyCost); err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "菜谱不存在"})
            return
        }
        c.JSON(http.StatusOK, menu)
    }
}

func UpdateMenu(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的菜谱 ID"})
            return
        }
        var menu models.Menu
        if err := c.ShouldBindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if menu.Name == "" || menu.EnergyCost <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "菜品名称和能量值不能为空或无效"})
            return
        }
        // 检查菜名是否与其他菜谱重复
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM menus WHERE name = ? AND id != ?", menu.Name, id).Scan(&count)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "检查菜名失败"})
            return
        }
        if count > 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "菜名已存在，请选择其他名称"})
            return
        }
        result, err := db.Exec("UPDATE menus SET name = ?, description = ?, energy_cost = ? WHERE id = ?",
            menu.Name, menu.Description, menu.EnergyCost, id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "更新菜谱失败"})
            return
        }
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "菜谱不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "菜谱更新成功"})
    }
}