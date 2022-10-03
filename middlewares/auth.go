package middlewares

import (
	"jwt-auth/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

// func Auth() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		tokenString := ctx.GetHeader("Authorization")
// 		if tokenString == "" {
// 			ctx.JSON(
// 				http.StatusUnauthorized,
// 				gin.H{"error": "request does not contain an access token"},
// 			)
// 			ctx.Abort()
// 			return
// 		}

// 		_, authErr := auth.ValidateToken(tokenString)
// 		if authErr.Err != nil {
// 			ctx.JSON(
// 				authErr.Status,
// 				gin.H{"error": authErr.Err.Error()},
// 			)
// 			ctx.Abort()
// 			return
// 		}

// 		ctx.Next()
// 	}
// }

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionId, err := ctx.Cookie("sessionId")
		if err != nil {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": "request does not contain an access token"},
			)
			ctx.Abort()
			return
		}

		err = auth.ValidateSession(sessionId)
		if err != nil {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": err.Error()},
			)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
