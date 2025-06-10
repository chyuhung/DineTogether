package handlers

import (
	"DineTogether/errors"
	"database/sql"
	"log"
	"net/http"
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
			c.Error(errors.NewAppError(400, "未加入任何 Party"))
			return
		}
		var order struct {
			MenuID int `json:"menu_id"`
		}
		if err := c.ShouldBindJSON(&order); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		var energyCost int
		row := db.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", order.MenuID)
		if err := row.Scan(&energyCost); err != nil {
			log.Printf("菜品不存在: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		var energyLeft int
		row = db.QueryRow("SELECT energy_left FROM parties WHERE id = ?", partyID)
		if err := row.Scan(&energyLeft); err != nil {
			log.Printf("Party 不存在: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		if energyLeft < energyCost {
			c.Error(errors.NewAppError(400, "Party 精力不足"))
			return
		}
		_, err := db.Exec("INSERT INTO orders (party_id, user_id, menu_id) VALUES (?, ?, ?)",
			partyID, userID, order.MenuID)
		if err != nil {
			log.Printf("点餐失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		_, err = db.Exec("UPDATE parties SET energy_left = energy_left - ? WHERE id = ?",
			energyCost, partyID)
		if err != nil {
			log.Printf("更新 Party 精力失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		log.Printf("用户 %v 在 Party %v 点餐 %v 成功", userID, partyID, order.MenuID)
		c.JSON(http.StatusOK, gin.H{"message": "点餐成功"})
	}
}

func DeleteOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		partyID := session.Get("party_id")
		if partyID == nil {
			c.Error(errors.NewAppError(400, "未加入任何 Party"))
			return
		}
		orderID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		var orderUserID, menuID int
		row := db.QueryRow("SELECT user_id, menu_id FROM orders WHERE id = ? AND party_id = ?", orderID, partyID)
		if err := row.Scan(&orderUserID, &menuID); err != nil {
			log.Printf("订单不存在: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		if orderUserID != userID {
			c.Error(errors.NewAppError(403, "无权限删除他人订单"))
			return
		}
		var energyCost int
		row = db.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", menuID)
		if err := row.Scan(&energyCost); err != nil {
			log.Printf("菜品不存在: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("开启事务失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		_, err = tx.Exec("DELETE FROM orders WHERE id = ?", orderID)
		if err != nil {
			tx.Rollback()
			log.Printf("删除订单失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		_, err = tx.Exec("UPDATE parties SET energy_left = energy_left + ? WHERE id = ?",
			energyCost, partyID)
		if err != nil {
			tx.Rollback()
			log.Printf("恢复 Party 精力失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("提交事务失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		log.Printf("用户 %v 在 Party %v 删除订单 %v 成功", userID, partyID, orderID)
		c.JSON(http.StatusOK, gin.H{"message": "订单删除成功"})
	}
}
