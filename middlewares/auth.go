package middlewares

import (
	"fmt"
	db "jwt-auth/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIAuth() gin.HandlerFunc {
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

		err = db.ValidateSession(ctx, sessionId)
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

func GlobeAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionId, err := ctx.Cookie("sessionId")
		if err != nil {
			ctx.Error(fmt.Errorf("request does not contain an access token NO COOKIE"))
			ctx.Next()
			return
		}

		err = db.ValidateSession(ctx, sessionId)
		if err != nil {
			ctx.Error(fmt.Errorf("request does not contain an access token NOT FOUND"))
		}

		ctx.Next()
	}
}
