package controllers

import (
	db "jwt-auth/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Logout(ctx *gin.Context) {
	sessionId := db.ExtractSessionId(ctx.Request.Cookies()[0])
	db.RemoveUserSession(ctx, sessionId)

	ctx.SetCookie("sessionId", sessionId, -1, "/", "localhost", true, true)
	ctx.Header("Location", "/")

	ctx.Redirect(http.StatusFound, "/")
}
