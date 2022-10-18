package main

import (
	"errors"
	"fmt"
	"html/template"
	"jwt-auth/config"
	env "jwt-auth/config"
	"jwt-auth/controllers"
	db "jwt-auth/database"
	"jwt-auth/middlewares"
	"jwt-auth/models"
	tempConf "jwt-auth/templates"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/contrib/static"
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
}

func loadTemplates() {
	tempConf.EmailTemplate = template.Must(template.ParseFiles("./templates/emailTemplate.html"))
}

func main() {
	loadEnv()
	loadTemplates()

	// Initialize databases
	db.MainDB.Connect("root:example@tcp(localhost:3306)/main_DB?parseTime=true", &gorm.Config{})
	db.MainDB.Migrate(&models.User{})
	db.SessionDB.Connect("root:example@tcp(localhost:3306)/session_DB?parseTime=true", &gorm.Config{})
	db.SessionDB.Migrate(&models.Session{})

	// Initialize router
	router := gin.Default()
	initRouter(router)
	router.Run(":" + config.PORT)
}

func StaticFiles() gin.HandlerFunc {
	fsLogin := static.LocalFile("./views/login/", true)
	fsDashboard := static.LocalFile("./views/dashboard/", true)

	fileserverL := http.FileServer(fsLogin)
	fileserverL = http.StripPrefix("/", fileserverL)

	fileserverD := http.FileServer(fsDashboard)
	fileserverD = http.StripPrefix("/", fileserverD)

	return func(c *gin.Context) {
		if c.Errors == nil {
			if fsDashboard.Exists("/", c.Request.URL.Path) {
				fileserverD.ServeHTTP(c.Writer, c.Request)
				c.Abort()
				return
			}
		}

		if fsLogin.Exists("/", c.Request.URL.Path) {
			fileserverL.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

func initRouter(router *gin.Engine) {
	// Do auth
	router.Use(middlewares.GlobeAuth())

	// React apps
	router.NoRoute(func(c *gin.Context) {
		view := "dashboard"
		if len(c.Errors) != 0 {
			view = "login"
			c.Header("Cache-Control", "no-store")
		}

		path := c.Request.URL.Path
		if path == "/" || path == "" || !strings.HasPrefix(path, "/static") {
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
	api.Use(middlewares.APIHeadersMiddleware())
	{
		api.POST("/login", controllers.Login)
		api.POST("/register", controllers.Register)
		api.GET("/activate/:token", controllers.Activate)
		secured := api.Group("/").Use(middlewares.APIAuth())
		{
			secured.GET("/ping", controllers.Ping)
		}
	}
}
