package handlers

import (
	"DineTogether/errors"
	"DineTogether/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMenus 获取所有菜品
func GetMenus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, description, energy_cost, image_urls FROM menus")
		if err != nil {
			log.Printf("查询菜品失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		defer rows.Close()
		var menus []models.Menu
		for rows.Next() {
			var menu models.Menu
			var description, imageURLs sql.NullString
			if err := rows.Scan(&menu.ID, &menu.Name, &description, &menu.EnergyCost, &imageURLs); err != nil {
				log.Printf("扫描菜品失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			menu.Description = description.String
			if imageURLs.Valid {
				if err := json.Unmarshal([]byte(imageURLs.String), &menu.ImageURLs); err != nil {
					log.Printf("解析 image_urls 失败: %v", err)
					c.Error(errors.ErrInternalServer)
					return
				}
			} else {
				menu.ImageURLs = []string{}
			}
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
		imageURLsJSON, err := json.Marshal(menu.ImageURLs)
		if err != nil {
			log.Printf("序列化 image_urls 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		result, err := db.Exec("INSERT INTO menus (name, description, energy_cost, image_urls) VALUES (?, ?, ?, ?)", menu.Name, menu.Description, menu.EnergyCost, string(imageURLsJSON))
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
		idStr := c.Param("id")
		if idStr == "undefined" || idStr == "" {
			c.Error(errors.NewAppError(http.StatusBadRequest, "无效的菜品 ID"))
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		var menu models.Menu
		var description, imageURLs sql.NullString
		row := db.QueryRow("SELECT id, name, description, energy_cost, image_urls FROM menus WHERE id = ?", id)
		if err := row.Scan(&menu.ID, &menu.Name, &description, &menu.EnergyCost, &imageURLs); err != nil {
			if err == sql.ErrNoRows {
				c.Error(errors.NewAppError(http.StatusNotFound, "菜品不存在"))
			} else {
				log.Printf("查询菜品失败: %v", err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		menu.Description = description.String
		if imageURLs.Valid {
			if err := json.Unmarshal([]byte(imageURLs.String), &menu.ImageURLs); err != nil {
				log.Printf("解析 image_urls 失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
		} else {
			menu.ImageURLs = []string{}
		}
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
		imageURLsJSON, err := json.Marshal(menu.ImageURLs)
		if err != nil {
			log.Printf("序列化 image_urls 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		result, err := db.Exec("UPDATE menus SET name = ?, description = ?, energy_cost = ?, image_urls = ? WHERE id = ?", menu.Name, menu.Description, menu.EnergyCost, string(imageURLsJSON), id)
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

// DeleteMenu 删除菜品并恢复相关 Party 精力
func DeleteMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开启事务失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		// 获取菜品的精力消耗
		var energyCost int
		row := tx.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", id)
		if err := row.Scan(&energyCost); err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				c.Error(errors.NewAppError(http.StatusNotFound, "菜品不存在"))
			} else {
				log.Printf("查询菜品 %v 失败: %v", id, err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		// 查找所有引用该菜品的订单及其对应的 Party
		rows, err := tx.Query("SELECT party_id FROM orders WHERE menu_id = ? AND menu_id != 0", id)
		if err != nil {
			tx.Rollback()
			log.Printf("查询订单失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		defer rows.Close()
		partyEnergyUpdates := make(map[int]int) // Party ID -> 总精力恢复量
		for rows.Next() {
			var partyID int
			if err := rows.Scan(&partyID); err != nil {
				tx.Rollback()
				log.Printf("扫描订单失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			partyEnergyUpdates[partyID] += energyCost
		}
		// 更新每个 Party 的精力值
		for partyID, energyToRestore := range partyEnergyUpdates {
			_, err := tx.Exec("UPDATE parties SET energy_left = energy_left + ? WHERE id = ?", energyToRestore, partyID)
			if err != nil {
				tx.Rollback()
				log.Printf("恢复 Party %v 精力失败: %v", partyID, err)
				c.Error(errors.ErrInternalServer)
				return
			}
		}
		// 删除订单中的引用
		_, err = tx.Exec("DELETE FROM orders WHERE menu_id = ?", id)
		if err != nil {
			tx.Rollback()
			log.Printf("删除订单引用失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		// 删除菜品
		result, err := tx.Exec("DELETE FROM menus WHERE id = ?", id)
		if err != nil {
			tx.Rollback()
			log.Printf("删除菜品 %v 失败: %v", id, err)
			c.Error(errors.ErrInternalServer)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			tx.Rollback()
			c.Error(errors.NewAppError(http.StatusNotFound, "菜品不存在"))
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("提交事务失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		log.Printf("菜品 %v 删除成功，已恢复相关 Party 精力", id)
		c.JSON(http.StatusOK, gin.H{"message": "菜品删除成功"})
	}
}
