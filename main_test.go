package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	env "jwt-auth/config"
	"jwt-auth/controllers"
	db "jwt-auth/database"
	"jwt-auth/models"
	jwt "jwt-auth/token"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	loadEnv()
	loadTemplates()

	// Init main DB
	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	// gormConfig := &gorm.Config{}
	db.MainDB.Instance = initDB("main_DB_test", gormConfig)
	db.MainDB.Migrate(&models.User{})
	seedMainDB()

	// Init session DB
	db.SessionDB.Instance = initDB("session_DB_test", gormConfig)
	db.SessionDB.Migrate(&db.Session{})
	seedSessionDB()

	// Init session cache
	db.SessionCache.Connect(fmt.Sprintf("redis://default:%s@localhost:6379/0", env.CACHE_PASS))
	seedSessionCache()

	// Init router
	router = gin.New()
	initRouter(router)
	exitVal := m.Run()
	os.Exit(exitVal)
}

var userPasswords = []string{
	"john123smitH987!",
	"kitty_Minny789$",
	"Crunchy_biscuit_yes?",
	"Brian_lees_pass_123#",
	"Yikc32nick_trick?",
}

var seedUsers = []models.User{
	{Username: "jsmith123", Email: "jsmith@gmail.com", Password: getHashedPassword(userPasswords[0])},
	{Username: "barbara.gilth", Email: "barb_gilth123@yahoo.com", Password: getHashedPassword(userPasswords[1])},
	{Username: "aidenArmstrong", Email: "arm123_strong321@outlook.com", Password: getHashedPassword(userPasswords[2])},
	{Username: ".brianlees12", Email: "brian23@gmail.com", Password: getHashedPassword(userPasswords[3])},
	{Username: "kelly.kneeling11", Email: "kkelly456@gmail.com", Password: getHashedPassword(userPasswords[4])},
}

var seedSessions = []db.Session{
	{UserId: 1, SessionId: genSession(1).SessionId},
	{UserId: 2, SessionId: genSession(2).SessionId},
}

func initDB(dbName string, config *gorm.Config) *gorm.DB {
	// Create a new test DB
	dsn := fmt.Sprintf("root:%s@tcp(localhost:3306)/?parseTime=true", env.MAINDB_PASS)
	instance, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		panic(err.Error())
	}

	if result := instance.Exec("DROP DATABASE IF EXISTS " + dbName + ";"); result.Error != nil {
		panic(result.Error)
	}

	if result := instance.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + ";"); result.Error != nil {
		panic(result.Error)
	}

	// Reestablish connection to the newly created DB
	dsn = fmt.Sprintf("root:%s@tcp(localhost:3306)/%s?parseTime=true", env.MAINDB_PASS, dbName)
	instance, err = gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		panic(err.Error())
	}

	return instance
}

func seedMainDB() {
	for _, user := range seedUsers {
		err := db.MainDB.Instance.Where("email = ?", user.Email).First(&models.User{}).Error
		if err != nil {
			db.MainDB.Instance.Create(&user)
		}
	}
}

func seedSessionDB() {
	for _, s := range seedSessions {
		err := db.SessionDB.Instance.Where("sessionId = ?", s.SessionId).First(&db.Session{}).Error
		if err != nil {
			db.SessionDB.Instance.Create(&s)
		}
	}
}

func seedSessionCache() {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	for _, s := range seedSessions {
		err := db.SaveCacheSession(ctx, &s)
		if err != nil {
			panic("Failed to save a session to cache")
		}
	}
}

func getHashedPassword(password string) string {
	hashedPassword, _ := models.HashPassword(password)
	return hashedPassword
}

func genSession(userId uint) db.Session {
	s, _ := db.GenerateSession(1)
	return s
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

	assert.Equal(t, http.StatusConflict, w.Code)
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

func TestLogin_Successful(t *testing.T) {
	userId := 4

	// Return login page

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Successful login should redirect to the dashboard

	loginForm := controllers.LoginData{
		Email:    seedUsers[userId].Email,
		Password: userPasswords[userId],
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(loginForm)
	postBody := bytes.NewBuffer(reqBodyBytes.Bytes())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/login", postBody)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	location, _ := w.Result().Location()
	setCookieHeader := w.Result().Header.Get("Set-Cookie")
	sessionId := strings.Split(strings.Split(setCookieHeader, ";")[0], "=")[1]
	sessionId = strings.Replace(sessionId, "%3D", "=", -1)

	assert.Equal(t, "/", location.Path)
	assert.Equal(t, 360000, w.Result().Cookies()[0].MaxAge)
	assert.Equal(t, "/", w.Result().Cookies()[0].Path)
	assert.Equal(t, "localhost", w.Result().Cookies()[0].Domain)
	assert.Equal(t, true, w.Result().Cookies()[0].Secure)
	assert.Equal(t, true, w.Result().Cookies()[0].HttpOnly)

	// Successful login should create a new session in sessionDB

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	session, _ := db.ReadSessionDB(ctx, sessionId)
	assert.Equal(t, uint(userId+1), session.UserId)

	// Successful login should create a new session in session cache

	cacheUserId, _ := db.ReadSessionCache(ctx, sessionId)
	assert.Equal(t, fmt.Sprint(userId+1), cacheUserId)

	// Or we can check this in one go

	err := db.ValidateSession(ctx, sessionId)
	assert.Nil(t, err)
}

func TestLogin_InvalidPassword(t *testing.T) {
	userId := 1
	loginForm := controllers.LoginData{
		Email:    seedUsers[userId].Email,
		Password: "testPassword",
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(loginForm)
	postBody := bytes.NewBuffer(reqBodyBytes.Bytes())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/login", postBody)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_NoUser(t *testing.T) {
	loginForm := controllers.LoginData{
		Email:    "test@gmail.com",
		Password: "testPassword",
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(loginForm)
	postBody := bytes.NewBuffer(reqBodyBytes.Bytes())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/login", postBody)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =========================================================================
// === Logout ==============================================================
// =========================================================================

func TestLogout(t *testing.T) {
	userSession := seedSessions[0]

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/logout", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Cookie", fmt.Sprintf("sessionId=%s", userSession.SessionId))
	router.ServeHTTP(w, req)

	// Check if redirected to the home page and cookie is deleted

	assert.Equal(t, -1, w.Result().Cookies()[0].MaxAge)
	assert.Equal(t, http.StatusFound, w.Code)

	// Successful logout should remove user session from sessionDB

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, err := db.ReadSessionDB(ctx, userSession.SessionId)
	assert.Equal(t, fmt.Errorf("no session found"), err)

	// Successful logout should remove user session from session cache

	_, err = db.ReadSessionCache(ctx, userSession.SessionId)
	assert.NotNil(t, err)

	// Or we can check this in one go

	err = db.ValidateSession(ctx, userSession.SessionId)
	assert.NotNil(t, err)
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
	postBody := bytes.NewBuffer(reqBodyBytes.Bytes())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/reset", postBody)
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
	postBody := bytes.NewBuffer(reqBodyBytes.Bytes())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/reset", postBody)
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
	postBody := bytes.NewBuffer(reqBodyBytes.Bytes())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/reset", postBody)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	router.ServeHTTP(w, req)

	// Server should return success
	assert.Equal(t, http.StatusOK, w.Code)

	// Generate JWT token
	token, _ := jwt.GenerateResetJWT(form.Email)

	// --- Get new password and complete reset ---
	var resetForm = controllers.ResetForm{
		Password:  "newPassword123!",
		Password2: "newPassword123!",
		Token:     token,
	}

	reqBodyBytes = new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(resetForm)
	postBody = bytes.NewBuffer(reqBodyBytes.Bytes())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/complete-reset", postBody)
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
