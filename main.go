package main

import (
	"DineTogether/handlers"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	dbPath := viper.GetString("database.path")
	secret := viper.GetString("session.secret")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{MaxAge: 86400, Path: "/"})
	r.Use(sessions.Sessions("session", store))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.POST("/register", handlers.Register(db))
	r.POST("/login", handlers.Login(db))
	r.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", nil)
	})
	r.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
	})
	r.GET("/api/check-party", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		row := db.QueryRow(`
			SELECT p.id, p.name 
			FROM parties p 
			JOIN orders o ON p.id = o.party_id 
			WHERE o.user_id = ?`, userID)
		var partyID int
		var partyName string
		if err := row.Scan(&partyID, &partyName); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, gin.H{"hasParty": false})
				return
			}
			log.Printf("查询用户 Party 失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 Party 失败"})
			return
		}
		session.Set("party_id", partyID)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"hasParty": true, "party_id": partyID, "party_name": partyName})
	})
	r.GET("/api/party", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		row := db.QueryRow(`
			SELECT p.id, p.name 
			FROM parties p 
			JOIN orders o ON p.id = o.party_id 
			WHERE o.user_id = ?`, userID)
		var partyID int
		var partyName string
		if err := row.Scan(&partyID, &partyName); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, gin.H{"hasParty": false})
				return
			}
			log.Printf("查询用户 Party 失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 Party 失败"})
			return
		}
		session.Set("party_id", partyID)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"hasParty": true, "party_id": partyID, "party_name": partyName})
	})
	r.POST("/change-password", handlers.ChangePassword(db))
	r.GET("/join-party", func(c *gin.Context) {
		c.HTML(http.StatusOK, "join_party.html", nil)
	})
	r.POST("/join-party", handlers.JoinParty(db))
	r.POST("/leave-party", handlers.LeaveParty(db))
	r.GET("/order", func(c *gin.Context) {
		c.HTML(http.StatusOK, "order.html", nil)
	})
	r.POST("/order", handlers.PlaceOrder(db))
	r.GET("/api/party-orders", handlers.GetPartyOrders(db))
	r.DELETE("/order/:id", handlers.DeleteOrder(db))

	// 管理员页面路由（移除 /admin 分组，使用 AuthMiddleware 控制权限）
	r.GET("/menu-manage", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "menu_manage.html", nil)
	})
	r.GET("/create-menu", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "create_menu.html", nil)
	})
	r.GET("/edit-menu", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "edit_menu.html", nil)
	})
	r.GET("/party-manage", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "party_manage.html", nil)
	})
	r.GET("/create-party", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "create_party.html", nil)
	})
	r.GET("/edit-party", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "edit_party.html", nil)
	})
	r.GET("/user-manage", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "user_manage.html", nil)
	})
	r.GET("/create-user", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "create_user.html", nil)
	})
	r.GET("/edit-user", handlers.AuthMiddleware(db), func(c *gin.Context) {
		c.HTML(http.StatusOK, "edit_user.html", nil)
	})

	// API 路由（保持不变）
	r.GET("/menus", handlers.GetMenus(db))
	r.POST("/menus", handlers.AuthMiddleware(db), handlers.CreateMenu(db))
	r.GET("/menu/:id", handlers.AuthMiddleware(db), handlers.GetMenu(db))
	r.PUT("/menu/:id", handlers.AuthMiddleware(db), handlers.UpdateMenu(db))
	r.DELETE("/menu/:id", handlers.AuthMiddleware(db), handlers.DeleteMenu(db))
	r.GET("/parties", handlers.AuthMiddleware(db), handlers.GetParties(db))
	r.POST("/parties", handlers.AuthMiddleware(db), handlers.CreateParty(db))
	r.GET("/party/:id", handlers.AuthMiddleware(db), handlers.GetPartyByID(db))
	r.PUT("/party/:id", handlers.AuthMiddleware(db), handlers.UpdateParty(db))
	r.DELETE("/party/:id", handlers.AuthMiddleware(db), handlers.DeleteParty(db))
	r.GET("/users", handlers.AuthMiddleware(db), handlers.GetUsers(db))
	r.POST("/users", handlers.AuthMiddleware(db), handlers.CreateUser(db))
	r.GET("/user/:id", handlers.AuthMiddleware(db), handlers.GetUserByID(db))
	r.PUT("/user/:id", handlers.AuthMiddleware(db), handlers.UpdateUser(db))
	r.DELETE("/user/:id", handlers.AuthMiddleware(db), handlers.DeleteUser(db))

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
