package main

import (
    "database/sql"
    "DineTogether/handlers"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "net/http"
)

func main() {
    db, err := sql.Open("sqlite3", "./db/dine_together.db")
    if err != nil {
        log.Fatal("Failed to open database:", err)
    }
    defer db.Close()

    initDatabase(db)

    r := gin.Default()

    store := cookie.NewStore([]byte("secret-key"))
    store.Options(sessions.Options{
        Path:     "/",
        MaxAge:   86400,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    })
    r.Use(sessions.Sessions("dine_session", store))

    r.Static("/static", "./static")
    r.LoadHTMLGlob("templates/*")

    loginRequired := func(c *gin.Context) {
        session := sessions.Default(c)
        userID := session.Get("user_id")
        if userID == nil {
            c.Redirect(http.StatusFound, "/login")
            c.Abort()
            return
        }
        c.Next()
    }

    r.GET("/", func(c *gin.Context) {
        session := sessions.Default(c)
        if session.Get("user_id") == nil {
            c.Redirect(http.StatusFound, "/login")
            c.Abort()
            return
        }
        c.HTML(http.StatusOK, "index.html", nil)
    })
    r.GET("/login", func(c *gin.Context) {
        session := sessions.Default(c)
        if session.Get("user_id") != nil {
            c.Redirect(http.StatusFound, "/")
            c.Abort()
            return
        }
        c.HTML(http.StatusOK, "login.html", nil)
    })
    r.GET("/logout", loginRequired, func(c *gin.Context) {
        session := sessions.Default(c)
        session.Clear()
        if err := session.Save(); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "退出失败"})
            return
        }
        c.Redirect(http.StatusFound, "/login")
    })
    r.POST("/leave-party", loginRequired, handlers.LeaveParty(db))
    r.GET("/register", func(c *gin.Context) {
        c.HTML(http.StatusOK, "register.html", nil)
    })
    r.GET("/menu-manage", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "menu_manage.html", nil)
    })
    r.GET("/party-manage", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "party_manage.html", nil)
    })
    r.GET("/user-manage", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "user_manage.html", nil)
    })
    r.GET("/create-menu", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "create_menu.html", nil)
    })
    r.GET("/edit-menu/:id", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "edit_menu.html", nil)
    })
    r.GET("/create-party", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "create_party.html", nil)
    })
    r.GET("/edit-party/:id", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "edit_party.html", nil)
    })
    r.GET("/create-user", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "create_user.html", nil)
    })
    r.GET("/edit-user/:id", loginRequired, handlers.AuthMiddleware(db), func(c *gin.Context) {
        c.HTML(http.StatusOK, "edit_user.html", nil)
    })
    r.GET("/join-party", loginRequired, func(c *gin.Context) {
        c.HTML(http.StatusOK, "join_party.html", nil)
    })
    r.GET("/order", loginRequired, func(c *gin.Context) {
        c.HTML(http.StatusOK, "order.html", nil)
    })
    r.POST("/register", handlers.Register(db))
    r.POST("/login", handlers.Login(db))
    r.POST("/menus", loginRequired, handlers.AuthMiddleware(db), handlers.CreateMenu(db))
    r.GET("/menus", handlers.GetMenus(db))
    r.GET("/menu/:id", loginRequired, handlers.AuthMiddleware(db), handlers.GetMenuByID(db))
    r.PUT("/menu/:id", loginRequired, handlers.AuthMiddleware(db), handlers.UpdateMenu(db))
    r.DELETE("/menu/:id", loginRequired, handlers.AuthMiddleware(db), handlers.DeleteMenu(db))
    r.POST("/parties", loginRequired, handlers.AuthMiddleware(db), handlers.CreateParty(db))
    r.GET("/parties", handlers.GetParties(db))
    r.GET("/party/:id", loginRequired, handlers.AuthMiddleware(db), handlers.GetPartyByID(db))
    r.PUT("/party/:id", loginRequired, handlers.AuthMiddleware(db), handlers.UpdateParty(db))
    r.DELETE("/party/:id", loginRequired, handlers.AuthMiddleware(db), handlers.DeleteParty(db))
    r.POST("/users", loginRequired, handlers.AuthMiddleware(db), handlers.CreateUser(db))
    r.GET("/users", loginRequired, handlers.AuthMiddleware(db), handlers.GetUsers(db))
    r.GET("/user/:id", loginRequired, handlers.AuthMiddleware(db), handlers.GetUserByID(db))
    r.PUT("/user/:id", loginRequired, handlers.AuthMiddleware(db), handlers.UpdateUser(db))
    r.DELETE("/user/:id", loginRequired, handlers.AuthMiddleware(db), handlers.DeleteUser(db))
    r.POST("/join-party", loginRequired, handlers.JoinParty(db))
    r.POST("/order", loginRequired, handlers.PlaceOrder(db))

    if err := r.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}

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
        log.Fatal("Failed to create tables:", err)
    }
}