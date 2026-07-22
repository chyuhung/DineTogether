package handlers

import (
	"DineTogether/models"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetMenus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, description, energy_cost, image_urls FROM menus")
		if err != nil {
			log.Printf("查询菜品失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		defer rows.Close()

		menus := make([]models.Menu, 0)
		for rows.Next() {
			var menu models.Menu
			var description, imageURLs sql.NullString
			if err := rows.Scan(&menu.ID, &menu.Name, &description, &menu.EnergyCost, &imageURLs); err != nil {
				log.Printf("扫描菜品失败: %v", err)
				serverError(c, "服务器错误")
				return
			}
			menu.Description = description.String
			if imageURLs.Valid {
				if err := json.Unmarshal([]byte(imageURLs.String), &menu.ImageURLs); err != nil {
					log.Printf("解析 image_urls 失败: %v", err)
					serverError(c, "服务器错误")
					return
				}
			} else {
				menu.ImageURLs = []string{}
			}
			menus = append(menus, menu)
		}
		if err := rows.Err(); err != nil {
			log.Printf("遍历菜品行失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		success(c, "获取菜品列表成功", gin.H{"menus": menus})
	}
}

func CreateMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu
		if err := c.ShouldBindJSON(&menu); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if menu.Name == "" || menu.EnergyCost <= 0 {
			badRequest(c, "菜品名称和精力消耗不能为空且精力消耗必须大于0")
			return
		}
		imageURLsJSON, err := json.Marshal(menu.ImageURLs)
		if err != nil {
			log.Printf("序列化 image_urls 失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		result, err := db.Exec("INSERT INTO menus (name, description, energy_cost, image_urls) VALUES (?, ?, ?, ?)", menu.Name, menu.Description, menu.EnergyCost, string(imageURLsJSON))
		if err != nil {
			log.Printf("创建菜品失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		id, _ := result.LastInsertId()
		success(c, "菜品创建成功", gin.H{"menu_id": id})
	}
}

func GetMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "undefined" || idStr == "" {
			badRequest(c, "无效的菜品 ID")
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		var menu models.Menu
		var description, imageURLs sql.NullString
		row := db.QueryRow("SELECT id, name, description, energy_cost, image_urls FROM menus WHERE id = ?", id)
		if err := row.Scan(&menu.ID, &menu.Name, &description, &menu.EnergyCost, &imageURLs); err != nil {
			if err == sql.ErrNoRows {
				notFound(c, "菜品不存在")
			} else {
				log.Printf("查询菜品失败: %v", err)
				serverError(c, "服务器错误")
			}
			return
		}
		menu.Description = description.String
		if imageURLs.Valid {
			if err := json.Unmarshal([]byte(imageURLs.String), &menu.ImageURLs); err != nil {
				log.Printf("解析 image_urls 失败: %v", err)
				serverError(c, "服务器错误")
				return
			}
		} else {
			menu.ImageURLs = []string{}
		}
		success(c, "获取菜品成功", gin.H{"menu": menu})
	}
}

func UpdateMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		var menu models.Menu
		if err := c.ShouldBindJSON(&menu); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if menu.Name == "" || menu.EnergyCost <= 0 {
			badRequest(c, "菜品名称和精力消耗不能为空且精力消耗必须大于0")
			return
		}
		imageURLsJSON, err := json.Marshal(menu.ImageURLs)
		if err != nil {
			log.Printf("序列化 image_urls 失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		result, err := db.Exec("UPDATE menus SET name = ?, description = ?, energy_cost = ?, image_urls = ? WHERE id = ?", menu.Name, menu.Description, menu.EnergyCost, string(imageURLsJSON), id)
		if err != nil {
			log.Printf("更新菜品失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "菜品不存在")
			return
		}
		success(c, "菜品更新成功")
	}
}

func DeleteMenu(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开启事务失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		defer tx.Rollback()

		var energyCost int
		row := tx.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", id)
		if err := row.Scan(&energyCost); err != nil {
			if err == sql.ErrNoRows {
				notFound(c, "菜品不存在")
			} else {
				log.Printf("查询菜品 %v 失败: %v", id, err)
				serverError(c, "服务器错误")
			}
			return
		}
		rows, err := tx.Query("SELECT party_id FROM orders WHERE menu_id = ?", id)
		if err != nil {
			log.Printf("查询订单失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		partyEnergyUpdates := make(map[int]int)
		for rows.Next() {
			var partyID int
			if err := rows.Scan(&partyID); err != nil {
				log.Printf("扫描订单失败: %v", err)
				rows.Close()
				serverError(c, "服务器错误")
				return
			}
			partyEnergyUpdates[partyID] += energyCost
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			log.Printf("遍历订单行失败: %v", err)
			serverError(c, "服务器错误")
			return
		}

		for partyID, energyToRestore := range partyEnergyUpdates {
			_, err := tx.Exec("UPDATE parties SET energy_left = energy_left + ? WHERE id = ?", energyToRestore, partyID)
			if err != nil {
				log.Printf("恢复 Party %v 精力失败: %v", partyID, err)
				serverError(c, "服务器错误")
				return
			}
		}
		_, err = tx.Exec("DELETE FROM orders WHERE menu_id = ?", id)
		if err != nil {
			log.Printf("删除订单引用失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		result, err := tx.Exec("DELETE FROM menus WHERE id = ?", id)
		if err != nil {
			log.Printf("删除菜品 %v 失败: %v", id, err)
			serverError(c, "服务器错误")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "菜品不存在")
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("提交事务失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		log.Printf("菜品 %v 删除成功，已恢复相关 Party 精力", id)
		success(c, "菜品删除成功")
	}
}
