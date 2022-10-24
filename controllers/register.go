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

// Username: only alphanumeric and .
// Password: at least one uppercase, one lowercase, one digit, and a special character (@$!%*#?&^_-)

type RegistrationForm struct {
	Username  string `json:"username" validate:"required,min=6,max=20"`
	Email     string `json:"email" validate:"required,email,max=40"`
	Password  string `json:"password" validate:"required,min=10,max=30,containsany=@$!%*#?&^_-"`
	Password2 string `json:"password2" validate:"required,eqfield=Password"`
}

func Register(ctx *gin.Context) {
	var form RegistrationForm

	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid sign up form"},
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
			gin.H{"error": "internal server error"},
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
			gin.H{"error": "invalid sign up form"},
		)
		ctx.Abort()
		return err
	}

	// Check if the email exists in the database
	err = db.MainDB.Instance.Where("email = ?", form.Email).First(&user).Error
	if err == nil {
		ctx.JSON(
			http.StatusConflict,
			gin.H{"error": "email ID already registered"},
		)
		ctx.Abort()
		return err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
		ctx.Abort()
		return err
	}

	return nil
}

func ValidateSignupInputs(form RegistrationForm) error {
	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		return err
	}

	// Validate username
	if err := models.ValidateUsername(form.Username); err != nil {
		return err
	}

	// Validate password
	if err := models.ValidatePassword(form.Password); err != nil {
		return err
	}

	return nil
}

func sendValidationEmail(ctx *gin.Context, name, email, pass string) {
	// Generate JWT token
	token, err := auth.GenerateJWT(name, email, pass)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
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
			gin.H{"error": "internal server error"},
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
			gin.H{"error": "internal server error"},
		)
		ctx.Abort()
		return
	}
}
