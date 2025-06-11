package handlers

import (
	"DineTogether/errors"
	"DineTogether/models"
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMenus 获取所有菜品
func GetMenus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, description, energy_cost FROM menus")
		if err != nil {
			log.Printf("查询菜品失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		defer rows.Close()
		var menus []models.Menu
		for rows.Next() {
			var menu models.Menu
			var description sql.NullString
			if err := rows.Scan(&menu.ID, &menu.Name, &description, &menu.EnergyCost); err != nil {
				log.Printf("扫描菜品失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			menu.Description = description.String
			menus = append(menus, menu)
		}
		c.JSON(http.StatusOK, gin.H{"message": "获取菜品列表成功", "menus": menus})
	}
}

// CreateMenu 创建新菜品
func CreateMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu
		if err := c.ShouldBindJSON(&menu); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if menu.Name == "" || menu.EnergyCost <= 0 {
			c.Error(errors.NewAppError(http.StatusBadRequest, "菜品名称和精力消耗不能为空且精力消耗必须大于0"))
			return
		}
		result, err := db.Exec("INSERT INTO menus (name, description, energy_cost) VALUES (?, ?, ?)", menu.Name, menu.Description, menu.EnergyCost)
		if err != nil {
			log.Printf("创建菜品失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusOK, gin.H{"message": "菜品创建成功", "menu_id": id})
	}
}

// GetMenu 获取单个菜品
func GetMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		var menu models.Menu
		var description sql.NullString
		row := db.QueryRow("SELECT id, name, description, energy_cost FROM menus WHERE id = ?", id)
		if err := row.Scan(&menu.ID, &menu.Name, &description, &menu.EnergyCost); err != nil {
			if err == sql.ErrNoRows {
				c.Error(errors.NewAppError(http.StatusNotFound, "菜品不存在"))
			} else {
				log.Printf("查询菜品失败: %v", err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		menu.Description = description.String
		c.JSON(http.StatusOK, gin.H{"message": "获取菜品成功", "menu": menu})
	}
}

// UpdateMenu 更新菜品
func UpdateMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		var menu models.Menu
		if err := c.ShouldBindJSON(&menu); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if menu.Name == "" || menu.EnergyCost <= 0 {
			c.Error(errors.NewAppError(http.StatusBadRequest, "菜品名称和精力消耗不能为空且精力消耗必须大于0"))
			return
		}
		result, err := db.Exec("UPDATE menus SET name = ?, description = ?, energy_cost = ? WHERE id = ?", menu.Name, menu.Description, menu.EnergyCost, id)
		if err != nil {
			log.Printf("更新菜品失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.Error(errors.NewAppError(http.StatusNotFound, "菜品不存在"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "菜品更新成功"})
	}
}

// DeleteMenu 删除菜品
func DeleteMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		result, err := db.Exec("DELETE FROM menus WHERE id = ?", id)
		if err != nil {
			log.Printf("删除菜品失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.Error(errors.NewAppError(http.StatusNotFound, "菜品不存在"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "菜品删除成功"})
	}
}
