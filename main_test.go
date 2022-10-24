package main

import (
	"bytes"
	"encoding/json"
	"jwt-auth/auth"
	"jwt-auth/controllers"
	db "jwt-auth/database"
	"jwt-auth/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	loadEnv()
	loadTemplates()

	// Initialize database
	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	db.MainDB.Connect("root:example@tcp(localhost:3306)/main_DB_test?parseTime=true", gormConfig)
	db.MainDB.Migrate(&models.User{})
	seedDB()

	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	initRouter(router)
	exitVal := m.Run()
	os.Exit(exitVal)
}

var seedUsers = []models.User{
	{Username: "jsmith123", Email: "jsmith@gmail.com", Password: "john123smitH987!"},
	{Username: "barbara.gilth", Email: "barb_gilth123@yahoo.com", Password: "kitty_Minny789$"},
	{Username: "aidenArmstrong", Email: "arm123_strong321@outlook.com", Password: "Crunchy_biscuit_yes?"},
	{Username: ".brianlees12", Email: "brian23@gmail.com", Password: "Brian_lees_pass_123#"},
	{Username: "kelly.kneeling11", Email: "kkelly456@gmail.com", Password: "Yikc32nick_trick?"},
}

func seedDB() {
	for _, user := range seedUsers {
		err := db.MainDB.Instance.Where("email = ?", user.Email).First(&models.User{}).Error
		if err != nil {
			db.MainDB.Instance.Create(&user)
		}
	}
}

func getUserData(email string) (models.User, error) {
	var user models.User
	err := db.MainDB.Instance.Where("email = ?", email).First(&user).Error
	if err == nil {
		return user, err
	}
	return user, nil
}

// =========================================================================
// === Registration ========================================================
// =========================================================================

func TestRegister_AddNewUser(t *testing.T) {
	var jsonData = []byte(`{
		"username":  "testName",
		"email":     "test@gmail.com",
		"password":  "aA123465789!",
		"password2": "aA123465789!"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegister_AddExistingUser(t *testing.T) {
	var jsonData = []byte(`{
		"username":  "testName",
		"email":     "jsmith@gmail.com",
		"password":  "aA123465789!",
		"password2": "aA123465789!"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_AddInvalidUser(t *testing.T) {
	var form = controllers.RegistrationForm{
		Username:  "testName; DROP TABLE users; ",
		Email:     "test@test.com",
		Password:  "aA1234567890!",
		Password2: "aA1234567890!",
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(form)
	reqBodyBytes.Bytes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(reqBodyBytes.Bytes()))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	// Invalid username
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// =========================================================================
// === Login ===============================================================
// =========================================================================

func TestLoginRoute(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginAPIRoute(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =========================================================================
// === Password Reset ======================================================
// =========================================================================

func TestPassReset_NoEmail(t *testing.T) {
	var form = controllers.InitResetForm{
		Email: "test@test.com",
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(form)
	reqBodyBytes.Bytes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/reset", bytes.NewBuffer(reqBodyBytes.Bytes()))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	// Password reset request for a non-existing email returns OK status
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPassReset_InitReset(t *testing.T) {
	var form = controllers.InitResetForm{
		Email: seedUsers[0].Email,
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(form)
	reqBodyBytes.Bytes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/reset", bytes.NewBuffer(reqBodyBytes.Bytes()))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPassReset_FullPasswordReset(t *testing.T) {
	// --- Initiate password reset ---
	var form = controllers.InitResetForm{
		Email: seedUsers[0].Email,
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(form)
	reqBodyBytes.Bytes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/reset", bytes.NewBuffer(reqBodyBytes.Bytes()))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	// Server should return success
	assert.Equal(t, http.StatusOK, w.Code)

	// Generate JWT token
	token, _ := auth.GenerateResetJWT(form.Email)

	// --- Get new password and complete reset ---
	var resetForm = controllers.ResetForm{
		Password:  "newPassword123!",
		Password2: "newPassword123!",
		Token:     token,
	}

	reqBodyBytes = new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(resetForm)
	reqBodyBytes.Bytes()

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/complete-reset", bytes.NewBuffer(reqBodyBytes.Bytes()))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	// Server should return success
	assert.Equal(t, http.StatusOK, w.Code)
	// Password should be changed
	user, err := getUserData(form.Email)
	if err != nil {
		t.Errorf("Failed to get user data")
	}
	assert.Equal(t, user.CheckPassword(resetForm.Password), nil)
}

// =========================================================================
// === Unit Tests ==========================================================
// =========================================================================

var form = controllers.RegistrationForm{
	Username:  "testName",
	Email:     "test@test.com",
	Password:  "aA1234567890!",
	Password2: "aA1234567890!",
}

func TestRegistrationValidation(t *testing.T) {
	form := controllers.RegistrationForm{
		Username:  "testName",
		Email:     "test@test.com",
		Password:  "aA1234567890!",
		Password2: "aA1234567890!",
	}

	assert.NoError(t, controllers.ValidateSignupInputs(form))
}

func TestRegistrationUsername(t *testing.T) {
	newForm := form

	// Too short
	newForm.Username = "abcde"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	// Too long
	newForm.Username = "zxcvbnmlklpkdyirjhnbq"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	// Invalid characters
	newForm.Username = "username,"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	newForm.Username = "username|"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	newForm.Username = "username/"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	newForm.Username = "username:;&$!%@#?*^=<>(){}[]"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))
}

func TestRegistrationPassword(t *testing.T) {
	newForm := form

	// Too short
	newForm.Password = "123456789"
	newForm.Password2 = "123456789"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	// Too long
	newForm.Password = "zxcvbnmlkjhgfdsaqwertyuiop12345"
	newForm.Password2 = "zxcvbnmlkjhgfdsaqwertyuiop12345"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	// Passwords do not match
	newForm.Password = "password987654321"
	newForm.Password2 = "password123456789"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	// Invalid characters
	newForm.Password = "password123456,"
	newForm.Password2 = "password123456,"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	newForm.Password = "password123456/"
	newForm.Password2 = "password123456/"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	newForm.Password = "password123456<>=(){}[]"
	newForm.Password2 = "password123456<>=(){}[]"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))
}

func TestRegistrationEmail(t *testing.T) {
	newForm := form

	// Too long
	newForm.Email = "zxcvbnmlkjhgfdsaqwerqwert123456@gmail.com"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	// Invalid email
	newForm.Email = "123@gmail@com"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	newForm.Email = "test,@gmail.com"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))

	newForm.Email = "test@gmail"
	assert.NotNil(t, controllers.ValidateSignupInputs(newForm))
}
