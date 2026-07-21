package main

import (
	"DineTogether/handlers"
	"DineTogether/middleware"
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
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

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建数据库目录失败: %v", err)
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}
	runMigrations(db)

	r := gin.Default()
	r.Use(middleware.ErrorHandler())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://127.0.0.1:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{
		MaxAge:   86400,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("session", store))

	rl := middleware.NewRateLimiter(10, time.Minute)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.POST("/register", middleware.RateLimitMiddleware(rl), handlers.Register(db))
	r.POST("/login", middleware.RateLimitMiddleware(rl), handlers.Login(db))
	r.POST("/logout", middleware.CSRFMiddleware(), handlers.Logout(db))
	r.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", nil)
	})
	r.GET("/change-password", func(c *gin.Context) {
		c.HTML(http.StatusOK, "change_password.html", nil)
	})
	r.POST("/change-password", middleware.CSRFMiddleware(), handlers.ChangePassword(db))
	r.GET("/join-party", func(c *gin.Context) {
		c.HTML(http.StatusOK, "join_party.html", nil)
	})
	r.POST("/join-party", middleware.CSRFMiddleware(), handlers.JoinParty(db))
	r.POST("/leave-party", middleware.CSRFMiddleware(), handlers.LeaveParty(db))
	r.GET("/order", func(c *gin.Context) {
		c.HTML(http.StatusOK, "order.html", nil)
	})
	r.POST("/order", middleware.CSRFMiddleware(), handlers.PlaceOrder(db))
	r.GET("/api/party", handlers.GetUserParty(db))
	r.GET("/api/party-orders", handlers.GetPartyOrders(db))
	r.DELETE("/order/:id", middleware.CSRFMiddleware(), handlers.DeleteOrder(db))
	r.GET("/menu-detail", func(c *gin.Context) {
		c.HTML(http.StatusOK, "menu_detail.html", nil)
	})
	r.GET("/api/csrf-token", handlers.GetCSRFToken())

	adminRoutes := r.Group("")
	adminRoutes.Use(handlers.AuthMiddleware(db))
	{
		adminRoutes.GET("/menu-manage", func(c *gin.Context) {
			c.HTML(http.StatusOK, "menu_manage.html", nil)
		})
		adminRoutes.GET("/create-menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "create_menu.html", nil)
		})
		adminRoutes.GET("/edit-menu", func(c *gin.Context) {
			c.HTML(http.StatusOK, "edit_menu.html", nil)
		})
		adminRoutes.GET("/party-manage", func(c *gin.Context) {
			c.HTML(http.StatusOK, "party_manage.html", nil)
		})
		adminRoutes.GET("/create-party", func(c *gin.Context) {
			c.HTML(http.StatusOK, "create_party.html", nil)
		})
		adminRoutes.GET("/edit-party", func(c *gin.Context) {
			c.HTML(http.StatusOK, "edit_party.html", nil)
		})
		adminRoutes.GET("/user-manage", func(c *gin.Context) {
			c.HTML(http.StatusOK, "user_manage.html", nil)
		})
		adminRoutes.GET("/create-user", func(c *gin.Context) {
			c.HTML(http.StatusOK, "create_user.html", nil)
		})
		adminRoutes.GET("/edit-user", func(c *gin.Context) {
			c.HTML(http.StatusOK, "edit_user.html", nil)
		})
		adminRoutes.POST("/menus", middleware.CSRFMiddleware(), handlers.CreateMenu(db))

		adminRoutes.POST("/upload-image", handlers.UploadImage())
		adminRoutes.POST("/delete-image", handlers.DeleteImage())
		adminRoutes.GET("/menus", handlers.GetMenus(db))
		adminRoutes.GET("/menu/:id", handlers.GetMenu(db))
		adminRoutes.PUT("/menu/:id", middleware.CSRFMiddleware(), handlers.UpdateMenu(db))
		adminRoutes.DELETE("/menu/:id", middleware.CSRFMiddleware(), handlers.DeleteMenu(db))
		adminRoutes.GET("/parties", handlers.GetParties(db))
		adminRoutes.POST("/parties", middleware.CSRFMiddleware(), handlers.CreateParty(db))
		adminRoutes.GET("/party/:id", handlers.GetPartyByID(db))
		adminRoutes.PUT("/party/:id", middleware.CSRFMiddleware(), handlers.UpdateParty(db))
		adminRoutes.DELETE("/party/:id", middleware.CSRFMiddleware(), handlers.DeleteParty(db))
		adminRoutes.GET("/users", handlers.GetUsers(db))
		adminRoutes.POST("/users", middleware.CSRFMiddleware(), handlers.CreateUser(db))
		adminRoutes.GET("/user/:id", handlers.GetUserByID(db))
		adminRoutes.PUT("/user/:id", middleware.CSRFMiddleware(), handlers.UpdateUser(db))
		adminRoutes.DELETE("/user/:id", middleware.CSRFMiddleware(), handlers.DeleteUser(db))
	}

	r.GET("/menus", handlers.GetMenus(db))
	r.GET("/menu/:id", handlers.GetMenu(db))

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func runMigrations(db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'guest'
	);
	CREATE TABLE IF NOT EXISTS menus (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		energy_cost INTEGER NOT NULL CHECK(energy_cost > 0),
		image_urls TEXT DEFAULT '[]'
	);
	CREATE TABLE IF NOT EXISTS parties (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		energy_left INTEGER NOT NULL CHECK(energy_left >= 0),
		is_active INTEGER NOT NULL DEFAULT 1
	);
	CREATE TABLE IF NOT EXISTS party_members (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		party_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(party_id, user_id),
		FOREIGN KEY (party_id) REFERENCES parties(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		party_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		menu_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (party_id) REFERENCES parties(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
	);`
	for _, stmt := range strings.Split(schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			if _, err := db.Exec(stmt); err != nil {
				log.Printf("执行迁移失败: %v", err)
			}
		}
	}
}
