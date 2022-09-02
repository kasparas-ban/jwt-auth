package middlewares

import (
	"jwt-auth/auth"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			context.JSON(
				401,
				gin.H{"error": "request does not contain an access token"},
			)
			context.Abort()
			return
		}

		_, authErr := auth.ValidateToken(tokenString)
		if authErr.Err != nil {
			context.JSON(
				401,
				gin.H{"error": authErr.Err.Error()},
			)
			context.Abort()
			return
		}

		context.Next()
	}
}
