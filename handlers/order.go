package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Order struct {
	ID      int `json:"id"`
	PartyID int `json:"party_id"`
	UserID  int `json:"user_id"`
	MenuID  int `json:"menu_id"`
}

func PlaceOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var order Order
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
			return
		}

		// 检查 Party 是否活跃
		var party Party
		row := db.QueryRow("SELECT energy_left, is_active FROM parties WHERE id = ?", order.PartyID)
		if err := row.Scan(&party.EnergyLeft, &party.IsActive); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Party 不存在"})
			return
		}
		if !party.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Party 已关闭"})
			return
		}

		// 获取菜品精力值
		var energyCost int
		row = db.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", order.MenuID)
		if err := row.Scan(&energyCost); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "菜品不存在"})
			return
		}

		// 检查精力值是否足够
		if party.EnergyLeft < energyCost {
			c.JSON(http.StatusForbidden, gin.H{"error": "精力值不足"})
			return
		}

		// 扣除精力值
		_, err := db.Exec("UPDATE parties SET energy_left = energy_left - ? WHERE id = ?", energyCost, order.PartyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新精力值失败"})
			return
		}

		// 插入订单
		_, err = db.Exec("INSERT INTO orders (party_id, user_id, menu_id) VALUES (?, ?, ?)",
			order.PartyID, order.UserID, order.MenuID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "订单创建失败"})
			return
		}

		// 检查精力值是否耗尽
		if party.EnergyLeft-energyCost <= 0 {
			_, err = db.Exec("UPDATE parties SET is_active = false WHERE id = ?", order.PartyID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "关闭 Party 失败"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "订单提交成功"})
	}
}
