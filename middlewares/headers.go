package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AppsHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

func APIHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For POST only `application/json` content-type is allowed
		if c.Request.Method == "POST" && strings.ToLower(c.Request.Header.Get("Content-Type")) != "application/json; charset=utf-8" {
			c.JSON(
				http.StatusUnsupportedMediaType,
				gin.H{"error": "not acceptable content-type"},
			)
			c.Abort()
		}

		c.Header("Content-Type", "application/json; charset=UTF-8")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "deny")
		c.Header("X-XSS-Protection", "0")
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Security-Policy", "default-src 'none' frame-ancestors 'none'; sandbox")
		c.Header("Server", "''")
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
