package controllers

import (
	"bytes"
	auth "jwt-auth/auth"
	env "jwt-auth/config"
	db "jwt-auth/database"
	"jwt-auth/models"
	tmplConfig "jwt-auth/templates"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"gopkg.in/gomail.v2"
)

type InitResetForm struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetForm struct {
	Password  string `json:"password" validate:"required,min=10,max=30,containsany=@$!%*#?&^_-"`
	Password2 string `json:"password2" validate:"required,eqfield=Password"`
	Token     string `json:"token"`
}

func InitiateReset(ctx *gin.Context) {
	var form InitResetForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.AbortWithStatus(200)
		return
	}

	// Validate email
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		ctx.AbortWithStatus(200)
		return
	}

	// Send password reset email
	sendResetEmail(ctx, form.Email)
}

func sendResetEmail(ctx *gin.Context, email string) {
	// Generate JWT token
	token, err := auth.GenerateResetJWT(email)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
		ctx.Abort()
		return
	}
	resetUrl := "http://localhost:3001/reset-form/" + token

	// Load the template
	tmpl := &bytes.Buffer{}
	t := tmplConfig.ResetEmailTemplate
	if err = t.Execute(tmpl, resetUrl); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
		ctx.Abort()
		return
	}

	// Form the email
	m := gomail.NewMessage()
	m.SetHeader("From", "blue.dot <"+env.EMAIL_USER+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset")
	m.SetBody("text/html", tmpl.String())

	d := gomail.NewDialer(
		env.HOST_SERVER,
		env.EMAIL_PORT,
		env.EMAIL_USER,
		env.EMAIL_PASS,
	)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
		ctx.Abort()
		return
	}
}

func CompleteReset(ctx *gin.Context) {
	var form ResetForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "invalid reset form"},
		)
		ctx.Abort()
		return
	}

	// Validate reset form
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "invalid reset form"},
		)
		ctx.Abort()
		return
	}

	// Validate JWT
	claims, err := auth.ValidateResetJWT(form.Token)
	if err.Err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Msg},
		)
		ctx.Abort()
		return
	}

	// Hash the password
	hashedPassword, newErr := models.HashPassword(form.Password)
	if newErr != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to change the password"},
		)
		ctx.Abort()
		return
	}

	// Update user password
	result := db.MainDB.Instance.Model(&models.User{}).Where("email = ?", claims.Email).Update("password", hashedPassword)
	if result.Error != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to change the password"},
		)
		ctx.Abort()
		return
	}
}
