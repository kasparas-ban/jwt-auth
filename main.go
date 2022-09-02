package main

import (
	"html/template"
	"jwt-auth/config"
	env "jwt-auth/config"
	"jwt-auth/controllers"
	"jwt-auth/database"
	"jwt-auth/middlewares"
	tempConf "jwt-auth/templates"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
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

	// Initialize database
	database.Connect("root:example@tcp(localhost:3306)/jwt_demo?parseTime=true")
	database.Migrate()

	// Initialize router
	router := initRouter()
	router.Run(":" + config.PORT)
}

func initRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.POST("/token", controllers.GenerateToken)
		api.POST("/register", controllers.Register)
		secured := api.Group("/secured").Use(middlewares.Auth())
		{
			secured.GET("/ping", controllers.Ping)
		}
	}
	router.GET("/activate/:token", controllers.Activate)
	router.GET("/login", controllers.Login)
	return router
}
