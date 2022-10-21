package controllers

import (
	"bytes"
	"errors"
	auth "jwt-auth/auth"
	env "jwt-auth/config"
	db "jwt-auth/database"
	"jwt-auth/models"
	tmplConfig "jwt-auth/templates"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

// Username: only unicode characters
// Password: at leaset one uppercase, one lowercase, one digit, and a special character (@$!%*#?&^_-)

type RegistrationForm struct {
	Username  string `json:"username" validate:"required,min=6,max=20,alphaunicode"`
	Email     string `json:"email" validate:"required,email,max=40"`
	Password  string `json:"password" validate:"required,min=10,max=30,containsany=@$!%*#?&^_-"`
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

	if err := validateRegistrationForm(ctx, form); err != nil {
		return
	}

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

func validateRegistrationForm(ctx *gin.Context, form RegistrationForm) error {
	var user models.User

	err := ValidateSignupInputs(form)
	if err != nil {
		ctx.JSON(
			http.StatusUnprocessableEntity,
			gin.H{"error": "Invalid registration form"},
		)
		ctx.Abort()
		return err
	}

	// Check if the email exists in the database
	err = db.MainDB.Instance.Where("email = ?", form.Email).First(&user).Error
	if err == nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Email ID already registered"},
		)
		ctx.Abort()
		return err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return err
	}

	return nil
}

func ValidateSignupInputs(form RegistrationForm) error {
	validate := validator.New()
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
	t := tmplConfig.SignUpEmailTemplate
	if err := t.Execute(tmpl, confirmUrl); err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		ctx.Abort()
		return
	}

	// Form the email
	m := gomail.NewMessage()
	m.SetHeader("From", "blue.dot <"+env.EMAIL_USER+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "SignUp Confirmation")
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
