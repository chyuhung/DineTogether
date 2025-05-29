package handlers

import (
    "database/sql"
    "DineTogether/models"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "net/http"
)

func CreateParty(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var party models.Party
        if err := c.ShouldBindJSON(&party); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        if party.Name == "" || party.Password == "" || party.EnergyLeft <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Party 名称、密码和初始精力值不能为空或无效"})
            return
        }
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(party.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
            return
        }
        result, err := db.Exec("INSERT INTO parties (name, password, energy_left, is_active) VALUES (?, ?, ?, ?)",
            party.Name, string(hashedPassword), party.EnergyLeft, true)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "创建 Party 失败"})
            return
        }
        id, _ := result.LastInsertId()
        c.JSON(http.StatusOK, gin.H{"message": "Party 创建成功", "party_id": id})
    }
}

func JoinParty(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var joinRequest struct {
            PartyName string `json:"party_name"`
            Password  string `json:"password"`
            UserID    int    `json:"user_id"`
        }
        if err := c.ShouldBindJSON(&joinRequest); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
            return
        }
        var party models.Party
        row := db.QueryRow("SELECT id, name, password, energy_left, is_active FROM parties WHERE name = ? AND is_active = ?", joinRequest.PartyName, true)
        if err := row.Scan(&party.ID, &party.Name, &party.Password, &party.EnergyLeft, &party.IsActive); err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Party 不存在或已关闭"})
            return
        }
        if err := bcrypt.CompareHashAndPassword([]byte(party.Password), []byte(joinRequest.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "加入 Party 成功", "party_id": party.ID})
    }
}