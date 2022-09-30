package main

import (
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

func initRouter(router *gin.Engine) {
	// Login/Sign-up frontend
	router.Use(static.Serve("/static", static.LocalFile("./views/login/static", true)))
	router.NoRoute(func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "./views/login/index.html")
		c.Abort()
	})

	// Login/Sign-up API
	api := router.Group("/api")
	api.Use(middlewares.HeadersMiddleware())
	{
		api.POST("/login", controllers.Login)
		api.POST("/register", controllers.Register)
		api.GET("/activate/:token", controllers.Activate)
		secured := api.Group("/").Use(middlewares.Auth())
		{
			secured.GET("/ping", controllers.Ping)
		}
	}
}
