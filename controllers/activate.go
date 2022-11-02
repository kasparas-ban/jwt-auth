package controllers

import (
	"errors"
	env "jwt-auth/config"
	db "jwt-auth/database"
	"jwt-auth/models"
	jwt "jwt-auth/token"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Activate(ctx *gin.Context) {
	token := ctx.Param("token")

	timeoutMsg := "?timeout=true"
	userExistsMsg := "?exists=true"
	activatedMsg := "?activated=true"
	errorMsg := "?error=true"

	// Validate the token
	claims, authErr := jwt.ValidateJWT(token)
	if authErr.Status == http.StatusInternalServerError {
		ctx.Redirect(http.StatusFound, "http://localhost:"+env.PORT+"/"+errorMsg)
		return
	}
	if authErr.Status == http.StatusUnauthorized {
		ctx.Redirect(http.StatusGone, "http://localhost:"+env.PORT+"/"+timeoutMsg)
		return
	}

	// Check if the email exists in the database
	var user models.User
	err := db.MainDB.Instance.Where("email = ?", claims.Email).First(&user).Error
	if err == nil {
		ctx.Redirect(http.StatusFound, "http://localhost:"+env.PORT+"/"+userExistsMsg)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Redirect(http.StatusFound, "http://localhost:"+env.PORT+"/"+errorMsg)
		return
	}

	// Add user to the database
	newUser := models.User{
		Username: claims.Username,
		Email:    claims.Email,
		Password: claims.HashPass,
	}
	if db.SaveUser(&newUser) != nil {
		ctx.Redirect(http.StatusFound, "http://localhost:"+env.PORT+"/"+errorMsg)
		return
	}

	// Redirect to login page
	ctx.Redirect(http.StatusFound, "http://localhost:"+env.PORT+"/"+activatedMsg)
}
