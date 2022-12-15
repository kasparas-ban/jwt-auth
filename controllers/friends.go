package controllers

import (
	db "jwt-auth/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AllFriends(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	allfriends, err := db.GetAllFriendships(userID)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, allfriends)
}
