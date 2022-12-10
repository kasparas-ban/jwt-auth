package controllers

import (
	"errors"
	"net/http"

	db "jwt-auth/database"
	"jwt-auth/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteAccountForm struct {
	Password string `json:"password"`
}

func DeleteAccount(ctx *gin.Context) {
	var form DeleteAccountForm
	var user models.User

	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Get userId from the cookie
	sessionId := db.ExtractSessionId(ctx.Request.Cookies()[0])
	session, err := db.ReadSessionDB(ctx, sessionId)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Get user record
	record := db.MainDB.Instance.Where("id = ?", session.UserId).First(&user)
	if errors.Is(record.Error, gorm.ErrRecordNotFound) {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "invalid credentials"},
		)
		ctx.Abort()
		return
	} else if record.Error != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": record.Error.Error()},
		)
		ctx.Abort()
		return
	}

	// Check if password is correct
	credentialError := user.CheckPassword(form.Password)
	if credentialError != nil {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "invalid credentials"},
		)
		ctx.Abort()
		return
	}

	// Delete user record
	record = db.MainDB.Instance.Where("id = ?", session.UserId).Delete(&user)
	if errors.Is(record.Error, gorm.ErrRecordNotFound) {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "invalid credentials"},
		)
		ctx.Abort()
		return
	} else if record.Error != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": record.Error.Error()},
		)
		ctx.Abort()
		return
	}

	// Logout user
	db.RemoveUserSession(ctx, sessionId)
	ctx.SetCookie("sessionId", sessionId, -1, "/", "localhost", true, true)
	ctx.Header("Location", "/")
	ctx.Redirect(http.StatusFound, "/")
}
