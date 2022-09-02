package controllers

import (
	"errors"
	auth "jwt-auth/auth"
	db "jwt-auth/database"
	"jwt-auth/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Activate(ctx *gin.Context) {
	token := ctx.Param("token")

	// Validate the token
	claims, err := auth.ValidateToken(token)
	if err != nil {
		ctx.JSON( // Need to redirect with /?timeout=true
			http.StatusGone,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
	}

	// Check if the email exists in the database
	var user models.User
	err = db.Instance.Where("email = ?", claims.Email).First(&user).Error
	if err == nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Email ID already registered"},
		)
		ctx.Abort()
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
	}

	// Add the user to the database
	newUser := models.User{
		Username: claims.Username,
		Email:    claims.Email,
		Password: claims.Password,
	}
	result := db.Instance.Create(&newUser)
	if result.Error != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": result.Error.Error()},
		)
		ctx.Abort()
	}

	// Redirect to login page
	ctx.Redirect(http.StatusFound, "http://localhost:3001/login")
}
