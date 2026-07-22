package handlers

import (
	"DineTogether/models"
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var party models.Party
		if err := c.ShouldBindJSON(&party); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if party.Name == "" || party.Password == "" || party.EnergyLeft <= 0 {
			badRequest(c, "Party 名称、密码和初始精力值不能为空或无效")
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(party.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码加密失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		result, err := db.Exec("INSERT INTO parties (name, password, energy_left, is_active) VALUES (?, ?, ?, ?)", party.Name, hashedPassword, party.EnergyLeft, true)
		if err != nil {
			if isUniqueConstraint(err) {
				badRequest(c, "Party 名称已存在")
			} else {
				log.Printf("创建 Party 失败: %v", err)
				serverError(c, "服务器错误")
			}
			return
		}
		id, _ := result.LastInsertId()
		success(c, "Party 创建成功", gin.H{"party_id": id})
	}
}

func GetParties(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, energy_left, is_active FROM parties")
		if err != nil {
			log.Printf("获取 Party 列表失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		defer rows.Close()

		parties := make([]models.Party, 0)
		for rows.Next() {
			var party models.Party
			if err := rows.Scan(&party.ID, &party.Name, &party.EnergyLeft, &party.IsActive); err != nil {
				log.Printf("扫描 Party 失败: %v", err)
				serverError(c, "服务器错误")
				return
			}
			parties = append(parties, party)
		}
		if err := rows.Err(); err != nil {
			log.Printf("遍历 Party 行失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		success(c, "获取 Party 列表成功", gin.H{"parties": parties})
	}
}

func GetPartyByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		var party models.Party
		row := db.QueryRow("SELECT id, name, energy_left, is_active FROM parties WHERE id = ?", id)
		if err := row.Scan(&party.ID, &party.Name, &party.EnergyLeft, &party.IsActive); err != nil {
			log.Printf("Party %v 不存在: %v", id, err)
			notFound(c, "资源未找到")
			return
		}
		success(c, "获取 Party 成功", gin.H{"party": party})
	}
}

func UpdateParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		var party models.Party
		if err := c.ShouldBindJSON(&party); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if party.Name == "" || party.EnergyLeft <= 0 {
			badRequest(c, "Party 名称和精力值不能为空或无效")
			return
		}
		var hashedPassword string
		if party.Password != "" {
			hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(party.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("密码加密失败: %v", err)
				serverError(c, "服务器错误")
				return
			}
			hashedPassword = string(hashedPasswordBytes)
		} else {
			row := db.QueryRow("SELECT password FROM parties WHERE id = ?", id)
			if err := row.Scan(&hashedPassword); err != nil {
				log.Printf("Party %v 不存在: %v", id, err)
				notFound(c, "资源未找到")
				return
			}
		}
		result, err := db.Exec("UPDATE parties SET name = ?, password = ?, energy_left = ?, is_active = ? WHERE id = ?", party.Name, hashedPassword, party.EnergyLeft, party.IsActive, id)
		if err != nil {
			if isUniqueConstraint(err) {
				badRequest(c, "Party 名称已存在")
			} else {
				log.Printf("更新 Party %v 失败: %v", id, err)
				serverError(c, "服务器错误")
			}
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "资源未找到")
			return
		}
		success(c, "Party 更新成功")
	}
}

func DeleteParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		result, err := db.Exec("DELETE FROM parties WHERE id = ?", id)
		if err != nil {
			log.Printf("删除 Party %v 失败: %v", id, err)
			serverError(c, "服务器错误")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound(c, "资源未找到")
			return
		}
		success(c, "Party 删除成功")
	}
}

func JoinParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var joinRequest struct {
			PartyName string `json:"party_name"`
			Password  string `json:"password"`
		}
		if err := c.ShouldBindJSON(&joinRequest); err != nil {
			badRequest(c, "无效的请求数据")
			return
		}
		if joinRequest.PartyName == "" || joinRequest.Password == "" {
			badRequest(c, "Party 名称和密码不能为空")
			return
		}
		session := sessions.Default(c)
		userID, ok := session.Get("user_id").(int)
		if !ok || userID <= 0 {
			unauthorized(c, "用户未登录")
			return
		}
		var party models.Party
		row := db.QueryRow("SELECT id, name, password, energy_left, is_active FROM parties WHERE name = ? AND is_active = ?", joinRequest.PartyName, true)
		if err := row.Scan(&party.ID, &party.Name, &party.Password, &party.EnergyLeft, &party.IsActive); err != nil {
			log.Printf("Party %s 不存在或已关闭: %v", joinRequest.PartyName, err)
			unauthorized(c, "Party 不存在或已关闭")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(party.Password), []byte(joinRequest.Password)); err != nil {
			log.Printf("Party %s 密码错误", joinRequest.PartyName)
			unauthorized(c, "Party 密码错误")
			return
		}
		session.Set("party_id", party.ID)
		if err := session.Save(); err != nil {
			log.Printf("保存 session 失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		_, err := db.Exec("INSERT OR IGNORE INTO party_members (party_id, user_id) VALUES (?, ?)", party.ID, userID)
		if err != nil {
			log.Printf("记录用户 %v 加入 Party %v 失败: %v", userID, party.ID, err)
			serverError(c, "服务器错误")
			return
		}
		_, err = db.Exec("INSERT OR IGNORE INTO orders (party_id, user_id, menu_id) VALUES (?, ?, 0)", party.ID, userID)
		if err != nil {
			log.Printf("记录用户 %v 加入 Party %v 的订单失败: %v", userID, party.ID, err)
			serverError(c, "服务器错误")
			return
		}
		log.Printf("用户 %v 加入 Party %v (%s) 成功", userID, party.ID, party.Name)
		success(c, "加入 Party 成功", gin.H{
			"party_id":   party.ID,
			"party_name": party.Name,
		})
	}
}

func LeaveParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		partyID := session.Get("party_id")
		if partyID == nil {
			badRequest(c, "未加入任何 Party")
			return
		}
		_, err := db.Exec("DELETE FROM party_members WHERE user_id = ? AND party_id = ?", userID, partyID)
		if err != nil {
			log.Printf("用户 %v 离开 Party %v 失败: %v", userID, partyID, err)
			serverError(c, "服务器错误")
			return
		}
		_, err = db.Exec("DELETE FROM orders WHERE user_id = ? AND party_id = ?", userID, partyID)
		if err != nil {
			log.Printf("用户 %v 删除 Party %v 订单失败: %v", userID, partyID, err)
			serverError(c, "服务器错误")
			return
		}
		session.Delete("party_id")
		if err := session.Save(); err != nil {
			log.Printf("保存 session 失败: %v", err)
			serverError(c, "服务器错误")
			return
		}
		log.Printf("用户 %v 离开 Party %v 成功", userID, partyID)
		success(c, "离开 Party 成功")
	}
}

func GetCurrentParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		partyID := session.Get("party_id")
		if partyID == nil {
			c.JSON(200, gin.H{"message": "未加入 Party", "hasParty": false})
			return
		}
		var party models.Party
		row := db.QueryRow("SELECT id, name, energy_left, is_active FROM parties WHERE id = ?", partyID)
		if err := row.Scan(&party.ID, &party.Name, &party.EnergyLeft, &party.IsActive); err != nil {
			log.Printf("获取 Party %v 信息失败: %v", partyID, err)
			c.JSON(200, gin.H{"message": "未加入 Party", "hasParty": false})
			return
		}
		c.JSON(200, gin.H{
			"message":    "获取 Party 成功",
			"hasParty":   true,
			"party_id":   party.ID,
			"party_name": party.Name,
		})
	}
}
