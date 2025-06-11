package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// CheckParty 检查用户是否加入 Party
func CheckParty(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		partyID := session.Get("party_id")
		userID := session.Get("user_id")
		if partyID != nil {
			log.Printf("用户 %v 已加入 Party %v", userID, partyID)
			c.JSON(http.StatusOK, gin.H{"message": "已加入 Party", "hasParty": true})
			return
		}
		log.Printf("用户 %v 未加入任何 Party", userID)
		c.JSON(http.StatusOK, gin.H{"message": "未加入 Party", "hasParty": false})
	}
}
