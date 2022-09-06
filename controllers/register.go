package controllers

import (
	"bytes"
	"errors"
	auth "jwt-auth/auth"
	env "jwt-auth/config"
	db "jwt-auth/database"
	"jwt-auth/models"
	tmplConfig "jwt-auth/templates"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type RegistrationForm struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(ctx *gin.Context) {
	var form RegistrationForm

	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	validateRegistrationForm(ctx, form.Email)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}
	form.Password = string(hashedPassword)

	sendValidationEmail(ctx, form.Username, form.Email, form.Password)
}

func validateRegistrationForm(ctx *gin.Context, email string) {
	var user models.User

	// Check if the email exists in the database
	err := db.Instance.Where("email = ?", email).First(&user).Error
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
}

func sendValidationEmail(ctx *gin.Context, name, email, pass string) {
	// Generate JWT token
	token, err := auth.GenerateJWT(name, email, pass)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
	}
	confirmUrl := "http://localhost:3001/api/activate/" + token

	// Load the template
	tmpl := &bytes.Buffer{}
	getEmailHtml(tmpl, confirmUrl)

	// Form the email
	m := gomail.NewMessage()
	m.SetHeader("From", "Auth-Server <"+env.EMAIL_USER+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Registration Confirmation")
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
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}
}

func getEmailHtml(tmpl *bytes.Buffer, confirmUrl string) {
	t := tmplConfig.EmailTemplate

	if err := t.Execute(tmpl, confirmUrl); err != nil {
		log.Println(err)
	}
}
