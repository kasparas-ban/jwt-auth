package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only `application/json` content-type is allowed
		if c.Request.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
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
