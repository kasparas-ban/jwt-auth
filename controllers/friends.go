package controllers

import (
	db "jwt-auth/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AllFriends(ctx *gin.Context) {
	userID := ctx.GetUint64("userID")
	fromParam, toParam := ctx.DefaultQuery("from", "0"), ctx.DefaultQuery("to", "50")
	from, err1 := strconv.ParseUint(fromParam, 10, 32)
	to, err2 := strconv.ParseUint(toParam, 10, 32)
	if err1 != nil || err2 != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
		ctx.Abort()
		return
	}

	allfriends, err := db.GetAllFriendships(userID, int(to), int(from))
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
