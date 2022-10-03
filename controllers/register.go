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
	"github.com/go-playground/validator"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

// Username invalid characters: `/\,|:;&$!%@#?*^=<>(){}[]
// Password invalid characters: /\,|<>=(){}[]

type RegistrationForm struct {
	Username  string `json:"username" validate:"required,min=6,max=20,excludesall=0x60/0x5C0x2C0x7C:;&$!%@#?*^=<>(){}[]"`
	Email     string `json:"email" validate:"required,email,max=40"`
	Password  string `json:"password" validate:"required,min=10,max=30,excludesall=\\/0x2C0x7C<>=(){}[]"`
	Password2 string `json:"password2" validate:"required,eqfield=Password"`
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

	validateRegistrationForm(ctx, form)

	// Hash the password
	hashedPassword, err := models.HashPassword(form.Password)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}
	form.Password = hashedPassword

	sendValidationEmail(ctx, form.Username, form.Email, form.Password)
}

func validateRegistrationForm(ctx *gin.Context, form RegistrationForm) {
	var user models.User

	// Check if the email exists in the database
	err := db.MainDB.Instance.Where("email = ?", form.Email).First(&user).Error
	if err == nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Email ID already registered"},
		)
		ctx.Abort()
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	err = ValidateInputs(form)
	if err != nil {
		ctx.JSON(
			http.StatusUnprocessableEntity,
			gin.H{"error": "Invalid registration form"},
		)
		ctx.Abort()
		return
	}
}

func ValidateInputs(form RegistrationForm) error {
	var validate *validator.Validate
	validate = validator.New()

	err := validate.Struct(form)
	return err
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
		return
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
