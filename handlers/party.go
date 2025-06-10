package handlers

import (
	"DineTogether/errors"
	"DineTogether/models"
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var party models.Party
		if err := c.ShouldBindJSON(&party); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if party.Name == "" || party.Password == "" || party.EnergyLeft <= 0 {
			c.Error(errors.NewAppError(400, "Party 名称、密码和初始精力值不能为空或无效"))
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(party.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		result, err := db.Exec("INSERT INTO parties (name, password, energy_left, is_active) VALUES (?, ?, ?, ?)",
			party.Name, string(hashedPassword), party.EnergyLeft, true)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: parties.name" {
				c.Error(errors.NewAppError(400, "Party 名称已存在"))
			} else {
				log.Printf("创建 Party 失败: %v", err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusOK, gin.H{"message": "Party 创建成功", "party_id": id})
	}
}

func GetParties(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, energy_left, is_active FROM parties")
		if err != nil {
			log.Printf("获取 Party 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		defer rows.Close()
		var parties []models.Party
		for rows.Next() {
			var party models.Party
			if err := rows.Scan(&party.ID, &party.Name, &party.EnergyLeft, &party.IsActive); err != nil {
				log.Printf("扫描 Party 失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			parties = append(parties, party)
		}
		c.JSON(http.StatusOK, parties)
	}
}

func GetPartyByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		var party models.Party
		row := db.QueryRow("SELECT id, name, energy_left, is_active FROM parties WHERE id = ?", id)
		if err := row.Scan(&party.ID, &party.Name, &party.EnergyLeft, &party.IsActive); err != nil {
			log.Printf("Party 不存在: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, party)
	}
}

func UpdateParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		var party models.Party
		if err := c.ShouldBindJSON(&party); err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		if party.Name == "" || party.EnergyLeft <= 0 {
			c.Error(errors.NewAppError(400, "Party 名称和精力值不能为空或无效"))
			return
		}
		var hashedPassword string
		if party.Password != "" {
			hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(party.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("密码加密失败: %v", err)
				c.Error(errors.ErrInternalServer)
				return
			}
			hashedPassword = string(hashedPasswordBytes)
		} else {
			row := db.QueryRow("SELECT password FROM parties WHERE id = ?", id)
			if err := row.Scan(&hashedPassword); err != nil {
				log.Printf("Party 不存在: %v", err)
				c.Error(errors.ErrNotFound)
				return
			}
		}
		result, err := db.Exec("UPDATE parties SET name = ?, password = ?, energy_left = ?, is_active = ? WHERE id = ?",
			party.Name, hashedPassword, party.EnergyLeft, party.IsActive, id)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: parties.name" {
				c.Error(errors.NewAppError(400, "Party 名称已存在"))
			} else {
				log.Printf("更新 Party 失败: %v", err)
				c.Error(errors.ErrInternalServer)
			}
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.Error(errors.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Party 更新成功"})
	}
}

func DeleteParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Error(errors.ErrBadRequest)
			return
		}
		result, err := db.Exec("DELETE FROM parties WHERE id = ?", id)
		if err != nil {
			log.Printf("删除 Party 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.Error(errors.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Party 删除成功"})
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
			c.Error(errors.ErrBadRequest)
			return
		}
		var party models.Party
		row := db.QueryRow("SELECT id, name, password, energy_left, is_active FROM parties WHERE name = ? AND is_active = ?", joinRequest.PartyName, true)
		if err := row.Scan(&party.ID, &party.Name, &party.Password, &party.EnergyLeft, &party.IsActive); err != nil {
			log.Printf("Party 不存在或已关闭: %v", err)
			c.Error(errors.ErrNotFound)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(party.Password), []byte(joinRequest.Password)); err != nil {
			c.Error(errors.ErrUnauthorized)
			return
		}
		session := sessions.Default(c)
		session.Set("party_id", party.ID)
		if err := session.Save(); err != nil {
			log.Printf("保存 session 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		_, err := db.Exec("INSERT OR IGNORE INTO orders (party_id, user_id, menu_id) VALUES (?, ?, 0)",
			party.ID, joinRequest.UserID)
		if err != nil {
			log.Printf("记录用户加入 Party 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		log.Printf("用户 %d 加入 Party %d (%s) 成功", joinRequest.UserID, party.ID, party.Name)
		c.JSON(http.StatusOK, gin.H{"message": "加入 Party 成功", "party_id": party.ID, "party_name": party.Name})
	}
}

func LeaveParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		partyID := session.Get("party_id")
		if partyID == nil {
			c.Error(errors.NewAppError(400, "未加入任何 Party"))
			return
		}
		_, err := db.Exec("DELETE FROM orders WHERE user_id = ? AND party_id = ?", userID, partyID)
		if err != nil {
			log.Printf("离开 Party 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		session.Delete("party_id")
		if err := session.Save(); err != nil {
			log.Printf("保存 session 失败: %v", err)
			c.Error(errors.ErrInternalServer)
			return
		}
		log.Printf("用户 %v 离开 Party %v 成功", userID, partyID)
		c.JSON(http.StatusOK, gin.H{"message": "已离开 Party"})
	}
}

func GetCurrentParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		partyID := session.Get("party_id")
		if partyID == nil {
			c.JSON(http.StatusOK, gin.H{"hasParty": false})
			return
		}
		var party models.Party
		row := db.QueryRow("SELECT id, name, energy_left, is_active FROM parties WHERE id = ?", partyID)
		if err := row.Scan(&party.ID, &party.Name, &party.EnergyLeft, &party.IsActive); err != nil {
			log.Printf("获取 Party 信息失败: %v", err)
			c.JSON(http.StatusOK, gin.H{"hasParty": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"hasParty": true, "party_id": party.ID, "party_name": party.Name})
	}
}
