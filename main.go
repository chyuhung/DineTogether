package main

import (
    "database/sql"
    "DineTogether/handlers"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    _ "github.com/mattn/go-sqlite3"
    "log"
)

func main() {
    // 初始化数据库
    db, err := sql.Open("sqlite3", "./db/dine_together.db")
    if err != nil {
        log.Fatal("Failed to open database:", err)
    }
    defer db.Close()

    // 初始化数据库表
    initDatabase(db)

    // 创建 Gin 路由
    r := gin.Default()

    // 配置 session 中间件
    store := cookie.NewStore([]byte("secret-key"))
    r.Use(sessions.Sessions("dine_session", store))

    // 加载静态文件和模板
    r.Static("/static", "./static")
    r.LoadHTMLGlob("templates/*")

    // 注册路由
    r.GET("/", func(c *gin.Context) {
        c.HTML(200, "index.html", nil)
    })
    r.GET("/login", func(c *gin.Context) {
        c.HTML(200, "login.html", nil)
    })
    r.GET("/register", func(c *gin.Context) {
        c.HTML(200, "register.html", nil)
    })
    r.GET("/menu", func(c *gin.Context) {
        c.HTML(200, "menu.html", nil)
    })
    r.GET("/join-party", func(c *gin.Context) {
        c.HTML(200, "join_party.html", nil)
    })
    r.GET("/order", func(c *gin.Context) {
        c.HTML(200, "order.html", nil)
    })
    r.POST("/register", handlers.Register(db))
    r.POST("/login", handlers.Login(db))
    r.POST("/menu", handlers.AuthMiddleware(db), handlers.CreateMenu(db))
    r.GET("/menus", handlers.GetMenus(db))
    r.POST("/party", handlers.AuthMiddleware(db), handlers.CreateParty(db))
    r.POST("/join-party", handlers.JoinParty(db))
    r.POST("/order", handlers.PlaceOrder(db))

    // 启动服务器
    if err := r.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}

func initDatabase(db *sql.DB) {
    // 创建用户表
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL,
            role TEXT NOT NULL
        );
    `)
    if err != nil {
        log.Fatal("Failed to create users table:", err)
    }

    // 创建菜谱表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS menus (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            energy_cost INTEGER NOT NULL
        );
    `)
    if err != nil {
        log.Fatal("Failed to create menus table:", err)
    }

    // 创建 party 表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS parties (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            password TEXT NOT NULL,
            energy_left INTEGER NOT NULL,
            is_active BOOLEAN NOT NULL
        );
    `)
    if err != nil {
        log.Fatal("Failed to create parties table:", err)
    }

    // 创建订单表
    _, err = db.Exec(`
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
        log.Fatal("Failed to create orders table:", err)
    }
}