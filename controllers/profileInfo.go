package controllers

import (
	db "jwt-auth/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProfileInfo(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	user, err := db.GetProfileInfo(userID)

	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, user)
}
