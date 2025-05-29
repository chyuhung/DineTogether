package handlers

import (
    "database/sql"
    "DineTogether/models"
    "github.com/gin-gonic/gin"
    "net/http"
)

func CreateParty(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var party models.Party
        if err := c.ShouldBindJSON(&party); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }

        // 仅限管理员权限
        userRole := c.GetString("role")
        if userRole != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
            return
        }

        // 默认能量值和活跃状态
        party.EnergyLeft = 100
        party.IsActive = true

        _, err := db.Exec("INSERT INTO parties (name, password, energy_left, is_active) VALUES (?, ?, ?, ?)",
            party.Name, party.Password, party.EnergyLeft, party.IsActive)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "创建 Party 失败"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Party 创建成功"})
    }
}

func JoinParty(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            PartyName string `json:"party_name"`
            Password  string `json:"password"`
            UserID    int    `json:"user_id"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }

        var party models.Party
        row := db.QueryRow("SELECT id, name, password, energy_left, is_active FROM parties WHERE name = ? AND password = ?",
            input.PartyName, input.Password)
        if err := row.Scan(&party.ID, &party.Name, &party.Password, &party.EnergyLeft, &party.IsActive); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Party 不存在或密码错误"})
            return
        }

        if !party.IsActive {
            c.JSON(http.StatusForbidden, gin.H{"error": "Party 已关闭"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message":     "加入 Party 成功",
            "party_id":    party.ID,
            "energy_left": party.EnergyLeft,
        })
    }
}