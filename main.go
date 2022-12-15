package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	env "jwt-auth/config"
	"jwt-auth/controllers"
	db "jwt-auth/database"
	m "jwt-auth/middlewares"

	// "jwt-auth/models"
	tempConf "jwt-auth/templates"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func initializeDBs(dev bool) {
	if dev {
		db.MainDB.Connect(fmt.Sprintf("root:%s@tcp(localhost:3306)/main_DB?parseTime=true", env.MAINDB_PASS), &gorm.Config{})
		// db.MainDB.Migrate(&models.User{})
		db.SessionDB.Connect(fmt.Sprintf("root:%s@tcp(localhost:3306)/session_DB?parseTime=true", env.MAINDB_PASS), &gorm.Config{})
		db.SessionDB.Migrate(&db.Session{})
		db.SessionCache.Connect(fmt.Sprintf("redis://default:%s@localhost:6379/0", env.CACHE_PASS))
	} else {
		gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
		db.MainDB.Connect(fmt.Sprintf("root:%s@tcp(main_DB:3306)/main_DB?parseTime=true", env.MAINDB_PASS), gormConfig)
		// db.MainDB.Migrate(&models.User{})
		db.SessionDB.Connect(fmt.Sprintf("root:%s@tcp(main_DB:3306)/session_DB?parseTime=true", env.MAINDB_PASS), gormConfig)
		// db.SessionDB.Migrate(&db.Session{})
		db.SessionCache.Connect(fmt.Sprintf("redis://default:%s@sessions_cache:6379/0", env.CACHE_PASS))
	}
}

func main() {
	loadEnv()
	loadTemplates()

	environment := flag.Bool("dev", false, "environment description")
	flag.Parse()
	initializeDBs(*environment)

	// seed_data.PopulateMainDB()

	// Initialize router
	if !*environment {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	initRouter(router)
	router.Run(":" + env.PORT)
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
	secured := api.Group("/").Use(m.APIHeadersMiddleware())
	{
		// Signup / login
		secured.POST("/login", controllers.Login)
		secured.GET("/logout", controllers.Logout)
		secured.POST("/register", controllers.Register)
		api.GET("/activate/:token", controllers.Activate)

		// Updating user info
		secured.POST("/init-reset", controllers.InitiateReset)
		secured.POST("/complete-reset", controllers.CompleteReset)
		secured.POST("/deleteAccount", controllers.DeleteAccount)

		// Getting user info
		secured.GET("/allFriends", controllers.AllFriends)
	}
}
