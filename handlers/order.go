package handlers

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func PlaceOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		partyID := session.Get("party_id")
		if partyID == nil {
			badRequest(c, "未加入任何 Party")
			return
		}
		var order struct {
			MenuID int `json:"menu_id"`
		}
		if err := c.ShouldBindJSON(&order); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if order.MenuID <= 0 {
			badRequest(c, "无效的菜品 ID")
			return
		}
		var energyCost int
		row := db.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", order.MenuID)
		if err := row.Scan(&energyCost); err != nil {
			log.Printf("菜品 %v 不存在: %v", order.MenuID, err)
			notFound(c, "资源未找到")
			return
		}
		var energyLeft int
		row = db.QueryRow("SELECT energy_left FROM parties WHERE id = ?", partyID)
		if err := row.Scan(&energyLeft); err != nil {
			log.Printf("Party %v 不存在: %v", partyID, err)
			notFound(c, "资源未找到")
			return
		}
		if energyLeft < energyCost {
			badRequest(c, "Party 精力不足")
			return
		}
		var isMember bool
		row = db.QueryRow("SELECT EXISTS(SELECT 1 FROM party_members WHERE party_id = ? AND user_id = ?)", partyID, userID)
		if err := row.Scan(&isMember); err != nil || !isMember {
			badRequest(c, "未加入此 Party")
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开启事务失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO orders (party_id, user_id, menu_id) VALUES (?, ?, ?)", partyID, userID, order.MenuID)
		if err != nil {
			log.Printf("点餐失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		_, err = tx.Exec("UPDATE parties SET energy_left = energy_left - ? WHERE id = ?", energyCost, partyID)
		if err != nil {
			log.Printf("更新 Party %v 精力失败: %v", partyID, err)
			serverError(c, "服务器错误")
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("提交事务失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		log.Printf("用户 %v 在 Party %v 点餐 %v 成功", userID, partyID, order.MenuID)
		success(c, "点餐成功")
	}
}

func DeleteOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		partyID := session.Get("party_id")
		if partyID == nil {
			badRequest(c, "未加入任何 Party")
			return
		}
		userID := session.Get("user_id")
		if userID == nil {
			unauthorized(c, "用户未登录")
			return
		}
		orderID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			badRequest(c, "无效的订单 ID")
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开启事务失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		defer tx.Rollback()

		var menuID, energyCost int
		row := tx.QueryRow(`
			SELECT o.menu_id, m.energy_cost
			FROM orders o
			JOIN menus m ON o.menu_id = m.id
			WHERE o.id = ? AND o.user_id = ? AND o.party_id = ?`, orderID, userID, partyID)
		if err := row.Scan(&menuID, &energyCost); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("订单 %v 不存在或无权限: %v", orderID, err)
				notFound(c, "订单不存在")
				return
			}
			log.Printf("查询订单失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		result, err := tx.Exec("DELETE FROM orders WHERE id = ?", orderID)
		if err != nil {
			log.Printf("删除订单 %v 失败: %v", orderID, err)
			serverError(c, "服务器错误")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "订单不存在")
			return
		}
		_, err = tx.Exec("UPDATE parties SET energy_left = energy_left + ? WHERE id = ?", energyCost, partyID)
		if err != nil {
			log.Printf("更新 Party %v 精力值失败: %v", partyID, err)
			serverError(c, "服务器错误")
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("提交事务失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		log.Printf("删除订单 %v 成功，Party %v 精力值增加 %v", orderID, partyID, energyCost)
		success(c, "订单删除成功")
	}
}

func GetUserParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		var partyID int
		var partyName string
		row := db.QueryRow(`
			SELECT p.id, p.name
			FROM party_members pm
			JOIN parties p ON pm.party_id = p.id
			WHERE pm.user_id = ?`, userID)
		if err := row.Scan(&partyID, &partyName); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(200, gin.H{"hasParty": false})
				return
			}
			log.Printf("查询用户 Party 失败: %v", err)
			serverError(c, "查询 Party 失败")
			return
		}
		session.Set("party_id", partyID)
		session.Save()
		c.JSON(200, gin.H{"hasParty": true, "party_id": partyID, "party_name": partyName})
	}
}
