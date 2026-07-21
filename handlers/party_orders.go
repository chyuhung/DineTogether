package handlers

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type OrderItem struct {
	ID         int      `json:"id"`
	Username   string   `json:"username"`
	MenuName   string   `json:"menu_name"`
	MenuID     int      `json:"menu_id"`
	ImageURLs  []string `json:"image_urls"`
	EnergyCost int      `json:"energy_cost"`
	Quantity   int      `json:"quantity"`
}

func GetPartyOrders(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		partyID := session.Get("party_id")
		if partyID == nil {
			badRequest(c, "未加入任何 Party")
			return
		}
		var energyLeft int
		row := db.QueryRow("SELECT energy_left FROM parties WHERE id = ?", partyID)
		if err := row.Scan(&energyLeft); err != nil {
			log.Printf("获取 Party %v 剩余精力失败: %v", partyID, err)
			serverError(c, "服务器错误")
			return
		}
		rows, err := db.Query(`
			SELECT MIN(o.id) as id, u.username, m.name, m.id, m.image_urls, m.energy_cost, COUNT(*) as quantity
			FROM orders o
			JOIN users u ON o.user_id = u.id
			JOIN menus m ON o.menu_id = m.id
			WHERE o.party_id = ?
			GROUP BY u.id, m.id`, partyID)
		if err != nil {
			log.Printf("获取 Party %v 订单失败: %v", partyID, err)
			serverError(c, "服务器错误")
			return
		}
		defer rows.Close()

		var orders []OrderItem
		for rows.Next() {
			var order OrderItem
			var imageURLs sql.NullString
			if err := rows.Scan(&order.ID, &order.Username, &order.MenuName, &order.MenuID, &imageURLs, &order.EnergyCost, &order.Quantity); err != nil {
				log.Printf("扫描订单失败: %v", err)
				serverError(c, "服务器错误")
				return
			}
			if imageURLs.Valid {
				if err := json.Unmarshal([]byte(imageURLs.String), &order.ImageURLs); err != nil {
					log.Printf("解析 image_urls 失败: %v", err)
					serverError(c, "服务器错误")
					return
				}
			} else {
				order.ImageURLs = []string{}
			}
			orders = append(orders, order)
		}
		if err := rows.Err(); err != nil {
			log.Printf("遍历订单行失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		log.Printf("获取 Party %v 的订单成功，数量: %d", partyID, len(orders))
		success(c, "获取订单成功", gin.H{
			"orders":      orders,
			"energy_left": energyLeft,
		})
	}
}
