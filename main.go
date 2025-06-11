package main

import (
	"DineTogether/handlers"
	"DineTogether/middleware"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

// 初始化配置文件
func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}
}

// 初始化数据库表
func initDatabase(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS menus (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			energy_cost INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS parties (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			energy_left INTEGER NOT NULL,
			is_active BOOLEAN NOT NULL
		);
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			party_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			menu_id INTEGER NOT NULL,
			FOREIGN KEY (party_id) REFERENCES parties(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (menu_id) REFERENCES menus(id)
		);
	`)
	if err != nil {
		log.Fatalf("创建数据库表失败: %v", err)
	}
}

func main() {
	// 初始化配置
	initConfig()

	// 连接数据库
	db, err := sql.Open("sqlite3", viper.GetString("database.path"))
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	// 初始化数据库
	initDatabase(db)

	// 初始化 Gin 引擎
	r := gin.Default()
	r.Use(middleware.ErrorHandler())

	// 配置 session
	store := cookie.NewStore([]byte(viper.GetString("session.secret")))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1天
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	r.Use(sessions.Sessions("session", store))

	// 静态文件和模板
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 登录验证中间件
	loginRequired := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("user_id") == nil {
			log.Printf("未登录用户尝试访问 %s", c.Request.URL.Path)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}

	// 公共路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.POST("/login", handlers.Login(db))
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.POST("/register", handlers.Register(db))

	// 需要登录的路由
	protected := r.Group("/", loginRequired)
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", nil)
		})
		protected.GET("/logout", func(c *gin.Context) {
			session := sessions.Default(c)
			session.Clear()
			if err := session.Save(); err != nil {
				log.Printf("退出失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "退出失败"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
		})
		protected.GET("/api/user", handlers.GetUserInfo(db))
		protected.GET("/api/party", handlers.GetCurrentParty(db))
		protected.GET("/api/check-party", handlers.CheckParty(db))
		protected.GET("/api/party-orders", handlers.GetPartyOrders(db))
		protected.POST("/leave-party", handlers.LeaveParty(db))
		protected.GET("/join-party", func(c *gin.Context) {
			c.HTML(http.StatusOK, "join_party.html", nil)
		})
		protected.POST("/join-party", handlers.JoinParty(db))
		protected.GET("/change-password", func(c *gin.Context) {
			c.HTML(http.StatusOK, "change_password.html", nil)
		})
		protected.POST("/change-password", handlers.ChangePassword(db))
		protected.GET("/order", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id")
			partyID := session.Get("party_id")
			if partyID == nil {
				log.Printf("用户 %v 未加入 Party，跳转到 /dashboard", userID)
				c.Redirect(http.StatusFound, "/dashboard")
				c.Abort()
				return
			}
			log.Printf("用户 %v 进入点餐页面，Party %v", userID, partyID)
			c.HTML(http.StatusOK, "order.html", nil)
		})
		protected.POST("/order", handlers.PlaceOrder(db))
		protected.DELETE("/order/:id", handlers.DeleteOrder(db))
	}

	// 需要管理员权限的路由
	admin := r.Group("/", loginRequired, handlers.AuthMiddleware(db))
	{
		admin.GET("/menu-manage", func(c *gin.Context) {
			c.HTML(http.StatusOK, "menu_manage.html", nil)
		})
		admin.GET("/party-manage", func(c *gin.Context) {
			c.HTML(http.StatusOK, "party_manage.html", nil)
		})
		admin.GET("/user-manage", func(c *gin.Context) {
			c.HTML(http.StatusOK, "user_manage.html", nil)
		})
		admin.GET("/create-menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "create_menu.html", nil)
		})
		admin.GET("/edit-menu/:id", func(c *gin.Context) {
			c.HTML(http.StatusOK, "edit_menu.html", nil)
		})
		admin.GET("/create-party", func(c *gin.Context) {
			c.HTML(http.StatusOK, "create_party.html", nil)
		})
		admin.GET("/edit-party/:id", func(c *gin.Context) {
			c.HTML(http.StatusOK, "edit_party.html", nil)
		})
		admin.GET("/create-user", func(c *gin.Context) {
			c.HTML(http.StatusOK, "create_user.html", nil)
		})
		admin.GET("/edit-user/:id", func(c *gin.Context) {
			c.HTML(http.StatusOK, "edit_user.html", nil)
		})
		admin.POST("/menus", handlers.CreateMenu(db))
		admin.GET("/menus", handlers.GetMenus(db))
		admin.GET("/menu/:id", handlers.GetMenuByID(db))
		admin.PUT("/menu/:id", handlers.UpdateMenu(db))
		admin.DELETE("/menu/:id", handlers.DeleteMenu(db))
		admin.POST("/parties", handlers.CreateParty(db))
		admin.GET("/parties", handlers.GetParties(db))
		admin.GET("/party/:id", handlers.GetPartyByID(db))
		admin.PUT("/party/:id", handlers.UpdateParty(db))
		admin.DELETE("/party/:id", handlers.DeleteParty(db))
		admin.POST("/users", handlers.CreateUser(db))
		admin.GET("/users", handlers.GetUsers(db))
		admin.GET("/user/:id", handlers.GetUserByID(db))
		admin.PUT("/user/:id", handlers.UpdateUser(db))
		admin.DELETE("/user/:id", handlers.DeleteUser(db))
	}

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
