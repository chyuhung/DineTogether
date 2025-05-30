package handlers

import (
    "database/sql"
    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
    "log"
    "net/http"
)

type Order struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    MenuName string `json:"menu_name"`
}

func GetPartyOrders(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        partyID := session.Get("party_id")
        if partyID == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "未加入任何 Party"})
            return
        }
        rows, err := db.Query(`
            SELECT o.id, u.username, m.name
            FROM orders o
            JOIN users u ON o.user_id = u.id
            JOIN menus m ON o.menu_id = m.id
            WHERE o.party_id = ? AND o.menu_id != 0`, partyID)
        if err != nil {
            log.Printf("获取 Party 订单失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单失败"})
            return
        }
        defer rows.Close()
        var orders []Order
        for rows.Next() {
            var order Order
            if err := rows.Scan(&order.ID, &order.Username, &order.MenuName); err != nil {
                log.Printf("扫描订单失败: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描订单失败"})
                return
            }
            orders = append(orders, order)
        }
        log.Printf("获取 Party %v 的订单成功，数量: %d", partyID, len(orders))
        c.JSON(http.StatusOK, orders)
    }
}