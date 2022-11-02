package controllers

import (
	db "jwt-auth/database"
	"jwt-auth/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

type LoginData struct {
	Email    string `json:"email" validate:"required,email,max=40"`
	Password string `json:"password" validate:"required,min=10,max=30"`
}

func Login(ctx *gin.Context) {
	var form LoginData
	var user models.User

	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	if err := validateLoginInputs(form); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid login form"},
		)
		ctx.Abort()
		return
	}

	// Check if email exists
	record := db.MainDB.Instance.Where("email = ?", form.Email).First(&user)
	if record.Error != nil {
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

	// Generate SessionID
	newSession, err := db.GenerateSession(user.ID)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Set cookies header
	ctx.SetCookie("sessionId", newSession.SessionId, 360000, "/", "localhost", true, true) // TODO: change this
	ctx.Header("Location", "/")

	// Save session to sessionDB and cache
	err = db.SaveSession(ctx, &newSession)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Redirect to dashboard page
	ctx.Redirect(http.StatusFound, "/")
}

func validateLoginInputs(form LoginData) error {
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		return err
	}
	return nil
}
