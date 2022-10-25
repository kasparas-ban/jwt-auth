package main

import (
	"errors"
	"fmt"
	"html/template"
	"jwt-auth/config"
	env "jwt-auth/config"
	"jwt-auth/controllers"
	db "jwt-auth/database"
	m "jwt-auth/middlewares"
	"jwt-auth/models"
	tempConf "jwt-auth/templates"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

func loadEnv() {
	env.PORT = os.Getenv("PORT")
	env.JWT_KEY = os.Getenv("JWT_KEY")
	env.HOST_SERVER = os.Getenv("HOST_SERVER")
	env.EMAIL_PORT = env.GetEnvAsInt("EMAIL_PORT", 465)
	env.EMAIL_DOMAIN = os.Getenv("EMAIL_DOMAIN")
	env.EMAIL_USER = os.Getenv("EMAIL_USER")
	env.EMAIL_PASS = os.Getenv("EMAIL_PASS")
	env.CACHE_PASS = os.Getenv("CACHE_PASS")
	env.MAINDB_PASS = os.Getenv("MAINDB_PASS")
}

func loadTemplates() {
	tempConf.SignUpEmailTemplate = template.Must(template.ParseFiles("./templates/signupEmailTemplate.html"))
	tempConf.ResetEmailTemplate = template.Must(template.ParseFiles("./templates/resetEmailTemplate.html"))
	tempConf.ResetSuccessEmailTemplate = template.Must(template.ParseFiles("./templates/resetSuccessEmailTemplate.html"))
}

func main() {
	loadEnv()
	loadTemplates()

	// Initialize databases and cache
	db.MainDB.Connect(fmt.Sprintf("root:%s@tcp(main_db_test:3306)/main_DB?parseTime=true", config.MAINDB_PASS), &gorm.Config{})
	db.MainDB.Migrate(&models.User{})
	db.SessionDB.Connect(fmt.Sprintf("root:%s@tcp(main_db_test:3306)/session_DB?parseTime=true", config.MAINDB_PASS), &gorm.Config{})
	db.SessionDB.Migrate(&db.Session{})
	db.SessionCache.Connect(fmt.Sprintf("redis://default:%s@localhost:6379/0", config.CACHE_PASS))

	// Initialize router
	router := gin.Default()
	initRouter(router)
	router.Run(":" + config.PORT)
}

func initRouter(router *gin.Engine) {
	// Do auth
	router.Use(m.GlobeAuth())
	// router.Use(m.CORSMiddleware()) // TODO: remove in prod

	// React apps
	router.NoRoute(func(c *gin.Context) {
		view := "dashboard"
		if len(c.Errors) != 0 {
			view = "login"
			c.Header("Cache-Control", "no-store")
		}

		path := c.Request.URL.Path
		if path == "/" || path == "" {
			// if path == "/" || path == "" || !strings.HasPrefix(path, "/static") {
			path = "/index.html"
		}

		filePath := fmt.Sprintf("./views/%s%s", view, path)
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			c.File(fmt.Sprintf("./views/%s%s", view, "/index.html"))
			c.Abort()
			return
		}

		c.File(fmt.Sprintf("./views/%s%s", view, path))
		c.Abort()
	})

	// API routes
	api := router.Group("/api")
	{
		api.POST("/login", m.APIHeadersMiddleware(), controllers.Login)
		api.POST("/register", m.APIHeadersMiddleware(), controllers.Register)
		api.GET("/activate/:token", controllers.Activate)
		api.POST("/init-reset", m.APIHeadersMiddleware(), controllers.InitiateReset)
		api.POST("/complete-reset", m.APIHeadersMiddleware(), controllers.CompleteReset)
		secured := api.Group("/").Use(m.APIAuth())
		{
			secured.GET("/ping", controllers.Ping)
		}
	}
}
