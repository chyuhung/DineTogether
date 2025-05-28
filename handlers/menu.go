package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Menu struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	EnergyCost  int    `json:"energy_cost"`
}

func CreateMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu Menu
		if err := c.ShouldBindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
			return
		}

		// 仅限 admin 角色
		userRole := c.GetString("role") // 假设通过中间件设置
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		}

		_, err := db.Exec("INSERT INTO menus (name, description, energy_cost) VALUES (?, ?, ?)",
			menu.Name, menu.Description, menu.EnergyCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建菜谱失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "菜谱创建成功"})
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

		var menus []Menu
		for rows.Next() {
			var menu Menu
			if err := rows.Scan(&menu.ID, &menu.Name, &menu.Description, &menu.EnergyCost); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描菜谱失败"})
				return
			}
			menus = append(menus, menu)
		}

		c.JSON(http.StatusOK, menus)
	}
}
