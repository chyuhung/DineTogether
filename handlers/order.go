package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// PlaceOrder 提交订单
func PlaceOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		partyID := session.Get("party_id")
		if partyID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未加入任何 Party", "success": false})
			return
		}
		var order struct {
			MenuID int `json:"menu_id"`
		}
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "success": false})
			return
		}
		if order.MenuID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的菜品 ID", "success": false})
			return
		}
		var energyCost int
		row := db.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", order.MenuID)
		if err := row.Scan(&energyCost); err != nil {
			log.Printf("菜品 %v 不存在: %v", order.MenuID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到", "success": false})
			return
		}
		var energyLeft int
		row = db.QueryRow("SELECT energy_left FROM parties WHERE id = ?", partyID)
		if err := row.Scan(&energyLeft); err != nil {
			log.Printf("Party %v 不存在: %v", partyID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "资源未找到", "success": false})
			return
		}
		if energyLeft < energyCost {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Party 精力不足", "success": false})
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开启事务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		_, err = tx.Exec("INSERT INTO orders (party_id, user_id, menu_id) VALUES (?, ?, ?)", partyID, userID, order.MenuID)
		if err != nil {
			tx.Rollback()
			log.Printf("点餐失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		_, err = tx.Exec("UPDATE parties SET energy_left = energy_left - ? WHERE id = ?", energyCost, partyID)
		if err != nil {
			tx.Rollback()
			log.Printf("更新 Party %v 精力失败: %v", partyID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("提交事务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		log.Printf("用户 %v 在 Party %v 点餐 %v 成功", userID, partyID, order.MenuID)
		c.JSON(http.StatusOK, gin.H{"message": "点餐成功"})
	}
}

// DeleteOrder 删除单个订单并更新 Party 精力值
func DeleteOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		partyID := session.Get("party_id")
		if partyID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未加入任何 Party", "success": false})
			return
		}
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录", "success": false})
			return
		}
		orderID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单 ID", "success": false})
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开启事务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		defer tx.Rollback()

		// 获取订单信息
		var menuID, energyCost int
		row := tx.QueryRow(`
			SELECT o.menu_id, m.energy_cost
			FROM orders o
			JOIN menus m ON o.menu_id = m.id
			WHERE o.id = ? AND o.user_id = ? AND o.party_id = ?`, orderID, userID, partyID)
		if err := row.Scan(&menuID, &energyCost); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("订单 %v 不存在或无权限: %v", orderID, err)
				c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在", "success": false})
				return
			}
			log.Printf("查询订单失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}

		// 删除单个订单
		result, err := tx.Exec("DELETE FROM orders WHERE id = ?", orderID)
		if err != nil {
			log.Printf("删除订单 %v 失败: %v", orderID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("订单 %v 不存在: %v", orderID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在", "success": false})
			return
		}

		// 更新 Party 剩余精力值
		_, err = tx.Exec("UPDATE parties SET energy_left = energy_left + ? WHERE id = ?", energyCost, partyID)
		if err != nil {
			log.Printf("更新 Party %v 精力值失败: %v", partyID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("提交事务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}

		log.Printf("删除订单 %v 成功，Party %v 精力值增加 %v", orderID, partyID, energyCost)
		c.JSON(http.StatusOK, gin.H{"message": "订单删除成功"})
	}
}
