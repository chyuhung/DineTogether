package handlers

import (
	"DineTogether/errors"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Order 订单结构体，用于 JSON 响应
type Order struct {
	ID         int      `json:"id"`
	Username   string   `json:"username"`
	MenuName   string   `json:"menu_name"`
	MenuID     int      `json:"menu_id"`
	ImageURLs  []string `json:"image_urls"`
	EnergyCost int      `json:"energy_cost"`
	Quantity   int      `json:"quantity"`
}

// GetPartyOrders 获取当前 Party 的订单
func GetPartyOrders(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		partyID := session.Get("party_id")
		if partyID == nil {
			c.Error(errors.NewAppError(http.StatusBadRequest, "未加入任何 Party"))
			return
		}
		// 获取 Party 的剩余精力值
		var energyLeft int
		row := db.QueryRow("SELECT energy_left FROM parties WHERE id = ?", partyID)
		if err := row.Scan(&energyLeft); err != nil {
			log.Printf("获取 Party %v 剩余精力失败: %v", partyID, err)
			c.Error(errors.ErrInternalServer)
			return
		}
		rows, err := db.Query(`
			SELECT MIN(o.id) as id, u.username, m.name, m.id, m.image_urls, m.energy_cost, COUNT(*) as quantity
			FROM orders o
			JOIN users u ON o.user_id = u.id
			JOIN menus m ON o.menu_id = m.id
			WHERE o.party_id = ? AND o.menu_id != 0
			GROUP BY u.id, m.id`, partyID)
		if err != nil {
			log.Printf("获取 Party %v 订单失败: %v", partyID, err)
			c.Error(errors.ErrInternalServer)
			return
		}
		defer rows.Close()
		var orders []Order
		for rows.Next() {
			var order Order
			var imageURLs sql.NullString
			if err := rows.Scan(&order.ID, &order.Username, &order.MenuName, &order.MenuID, &imageURLs, &order.EnergyCost, &order.Quantity); err != nil {
				log.Printf("扫描订单失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			if imageURLs.Valid {
				if err := json.Unmarshal([]byte(imageURLs.String), &order.ImageURLs); err != nil {
					log.Printf("解析 image_urls 失败: %v", err)
					c.Error(errors.ErrInternalServer)
					return
				}
			} else {
				order.ImageURLs = []string{}
			}
			orders = append(orders, order)
		}
		log.Printf("获取 Party %v 的订单成功，数量: %d", partyID, len(orders))
		c.JSON(http.StatusOK, gin.H{
			"message":     "获取订单成功",
			"orders":      orders,
			"energy_left": energyLeft,
		})
	}
}
