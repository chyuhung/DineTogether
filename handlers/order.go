package handlers

import (
    "database/sql"
    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
    "log"
    "net/http"
)

func PlaceOrder(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        userID := session.Get("user_id")
        partyID := session.Get("party_id")
        if partyID == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "未加入任何 Party"})
            return
        }
        var order struct {
            MenuID int `json:"menu_id"`
        }
        if err := c.ShouldBindJSON(&order); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        var energyCost int
        row := db.QueryRow("SELECT energy_cost FROM menus WHERE id = ?", order.MenuID)
        if err := row.Scan(&energyCost); err != nil {
            log.Printf("菜品不存在: %v", err)
            c.JSON(http.StatusNotFound, gin.H{"error": "菜品不存在"})
            return
        }
        var energyLeft int
        row = db.QueryRow("SELECT energy_left FROM parties WHERE id = ?", partyID)
        if err := row.Scan(&energyLeft); err != nil {
            log.Printf("Party 不存在: %v", err)
            c.JSON(http.StatusNotFound, gin.H{"error": "Party 不存在"})
            return
        }
        if energyLeft < energyCost {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Party 精力不足"})
            return
        }
        _, err := db.Exec("INSERT INTO orders (party_id, user_id, menu_id) VALUES (?, ?, ?)",
            partyID, userID, order.MenuID)
        if err != nil {
            log.Printf("点餐失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "点餐失败"})
            return
        }
        _, err = db.Exec("UPDATE parties SET energy_left = energy_left - ? WHERE id = ?",
            energyCost, partyID)
        if err != nil {
            log.Printf("更新 Party 精力失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "更新 Party 精力失败"})
            return
        }
        log.Printf("用户 %v 在 Party %v 点餐 %v 成功", userID, partyID, order.MenuID)
        c.JSON(http.StatusOK, gin.H{"message": "点餐成功"})
    }
}