package main

import (
	"jwt-auth/controllers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginRoute(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := initRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// assert.Equal(t, "pong", w.Body.String())
}

func TestLoginAPIRoute(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := initRouter()

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
