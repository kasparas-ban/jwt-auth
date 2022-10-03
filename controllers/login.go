package controllers

import (
	auth "jwt-auth/auth"
	env "jwt-auth/config"
	db "jwt-auth/database"
	"jwt-auth/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(ctx *gin.Context) {
	var request LoginData
	var user models.User

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Check if email exists
	record := db.MainDB.Instance.Where("email = ?", request.Email).First(&user)
	if record.Error != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": record.Error.Error()},
		)
		ctx.Abort()
		return
	}

	// Check if password is correct
	credentialError := user.CheckPassword(request.Password)
	if credentialError != nil {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "invalid credentials"},
		)
		ctx.Abort()
		return
	}

	// Generate SessionID
	newSession, err := auth.GenerateSession(user.ID)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Set cookies header
	ctx.SetCookie("sessionId", newSession.SessionId, 3600, "/", "localhost", false, true) // TODO: change this

	// Add session to the session database
	err = auth.SaveSession(newSession)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Redirect to dashboard page
	ctx.Redirect(http.StatusFound, "http://localhost:"+env.PORT+"/dashboard")
}
