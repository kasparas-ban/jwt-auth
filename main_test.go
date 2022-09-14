package main

import (
	"bytes"
	"jwt-auth/controllers"
	"jwt-auth/database"
	"jwt-auth/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	loadEnv()
	loadTemplates()

	// Initialize database
	database.Connect("root:example@tcp(localhost:3306)/jwt_auth_DB_test?parseTime=true")
	database.Migrate()
	seedDB()

	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	initRouter(router)
	exitVal := m.Run()
	os.Exit(exitVal)
}

func seedDB() {
	users := []models.User{
		{Username: "jsmith123", Email: "jsmith@gmail.com", Password: "john123smith987!"},
		{Username: "barbara_gilth", Email: "barb_gilth123@yahoo.com", Password: "kitty_minny789$"},
		{Username: "aidenArmstrong32", Email: "arm123_strong321@outlook.com", Password: "crunchy_biscuit_yes?"},
		{Username: "_brianlees_", Email: "brian23@gmail.com", Password: "brian_lees_pass_123"},
		{Username: "kelly-kneeling", Email: "kkelly456@gmail.com", Password: "tikc32nick_trick?"},
	}
	for _, user := range users {
		err := database.Instance.Where("email = ?", user.Email).First(&models.User{}).Error
		if err != nil {
			database.Instance.Create(&user)
		}
	}
}

func TestRegister_AddNewUser(t *testing.T) {
	var jsonData = []byte(`{
		"username":  "test123",
		"email":     "test@gmail.com",
		"password":  "0123465789",
		"password2": "0123465789"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(jsonData))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegister_AddExistingUser(t *testing.T) {
	var jsonData = []byte(`{
		"username":  "test123",
		"email":     "jsmith@gmail.com",
		"password":  "0123465789",
		"password2": "0123465789"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_AddInvalidUser(t *testing.T) {
	var jsonData = []byte(`{
		"username":  "test123; DROP TABLE users; ",
		"email":     "test.username@gmail.com",
		"password":  "01234567890!",
		"password2":  "01234567890!",
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(jsonData))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginRoute(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// assert.Equal(t, "pong", w.Body.String())
}

func TestLoginAPIRoute(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/login", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// assert.Equal(t, "pong", w.Body.String())
}

var form = controllers.RegistrationForm{
	Username:  "test123",
	Email:     "test@test.com",
	Password:  "1234567890",
	Password2: "1234567890",
}

func TestRegistrationValidation(t *testing.T) {
	form := controllers.RegistrationForm{
		Username:  "test123",
		Email:     "test@test.com",
		Password:  "1234567890a",
		Password2: "1234567890a",
	}

	assert.NoError(t, controllers.ValidateInputs(form))
}

func TestRegistrationUsername(t *testing.T) {
	newForm := form

	// Too short
	newForm.Username = "abcde"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	// Too long
	newForm.Username = "zxcvbnmlklp1234567890"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	// Invalid characters
	newForm.Username = "username,"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	newForm.Username = "username|"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	newForm.Username = "username/"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	newForm.Username = "username:;&$!%@#?*^=<>(){}[]"
	assert.NotNil(t, controllers.ValidateInputs(newForm))
}

func TestRegistrationPassword(t *testing.T) {
	newForm := form

	// Too short
	newForm.Password = "123456789"
	newForm.Password2 = "123456789"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	// Too long
	newForm.Password = "zxcvbnmlkjhgfdsaqwertyuiop12345"
	newForm.Password2 = "zxcvbnmlkjhgfdsaqwertyuiop12345"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	// Passwords do not match
	newForm.Password = "password987654321"
	newForm.Password2 = "password123456789"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	// Invalid characters
	newForm.Password = "password123456,"
	newForm.Password2 = "password123456,"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	newForm.Password = "password123456/"
	newForm.Password2 = "password123456/"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	newForm.Password = "password123456<>=(){}[]"
	newForm.Password2 = "password123456<>=(){}[]"
	assert.NotNil(t, controllers.ValidateInputs(newForm))
}

func TestRegistrationEmail(t *testing.T) {
	newForm := form

	// Too long
	newForm.Email = "zxcvbnmlkjhgfdsaqwerqwert123456@gmail.com"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	// Invalid email
	newForm.Email = "123@gmail@com"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	newForm.Email = "test,@gmail.com"
	assert.NotNil(t, controllers.ValidateInputs(newForm))

	newForm.Email = "test@gmail"
	assert.NotNil(t, controllers.ValidateInputs(newForm))
}
