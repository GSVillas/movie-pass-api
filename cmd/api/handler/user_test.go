package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GSVillas/movie-pass-api/domain"
	"github.com/GSVillas/movie-pass-api/mock"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_Create_WhenJSONCannotBeDecoded_ShouldReturnUnprocessableEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{"status":422,"title":"Unable to Process Request","details":"We encountered an issue while trying to process your request. The data you provided is not in the expected format.","errors":[{"field":"payload","message":"The information provided is not correctly formatted or is missing required fields. Please review and try again."}]}`, rec.Body.String())
	}
}

func TestUserHandler_Create_WhenValidationFails_ShouldReturnUnprocessableEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"firstName": "",
		"lastName": "dsadsadsadsasa",
		"email": "test@example.com",
		"confirmEmail": "test@example.com",
		"password": "Str0ngP@ssw0rd!",
		"confirmPassword": "Str0ngP@ssw0rd!",
		"birthDate": "1990-01-01T00:00:00Z"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{
			"status": 422,
			"title": "Validation Error",
			"details": "One or more fields are invalid.",
			"errors": [
				{
					"field": "firstname",
					"message": "This field is required"
				}
			]
		}`, rec.Body.String())
	}
}

func TestUserPayload_Validate_WhenPasswordIsWeak_ShouldReturnValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"firstName": "John",
		"lastName": "Doe",
		"email": "test@example.com",
		"confirmEmail": "test@example.com",
		"password": "weakpass",
		"confirmPassword": "weakpass",
		"birthDate": "1990-01-01T00:00:00Z"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{
			"status": 422,
			"title": "Validation Error",
			"details": "One or more fields are invalid.",
			"errors": [
				{
					"field": "password",
					"message": "Password must be at least 8 characters long, contain an uppercase letter, a number, and a special character"
				}
			]
		}`, rec.Body.String())
	}
}

func TestUserPayload_Validate_WhenUserIsTooOld_ShouldReturnValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"firstName": "John",
		"lastName": "Doe",
		"email": "test@example.com",
		"confirmEmail": "test@example.com",
		"password": "Str0ngP@ssw0rd!",
		"confirmPassword": "Str0ngP@ssw0rd!",
		"birthDate": "1820-01-01T00:00:00Z"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{
			"status": 422,
			"title": "Validation Error",
			"details": "One or more fields are invalid.",
			"errors": [
				{
					"field": "birthdate",
					"message": "The date of birth indicates an age greater than the allowed maximum of 200 years"
				}
			]
		}`, rec.Body.String())
	}
}

func TestUserPayload_Validate_WhenBirthDateIsInFuture_ShouldReturnValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"firstName": "John",
		"lastName": "Doe",
		"email": "test@example.com",
		"confirmEmail": "test@example.com",
		"password": "Str0ngP@ssw0rd!",
		"confirmPassword": "Str0ngP@ssw0rd!",
		"birthDate": "2100-01-01T00:00:00Z"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{
			"status": 422,
			"title": "Validation Error",
			"details": "One or more fields are invalid.",
			"errors": [
				{
					"field": "birthdate",
					"message": "The date of birth cannot be in the future"
				}
			]
		}`, rec.Body.String())
	}
}

func TestUserHandler_Create_WhenInternalServerErrorOccurs_ShouldReturnInternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"firstName": "John",
		"lastName": "Doe",
		"email": "test@example.com",
		"confirmEmail": "test@example.com",
		"password": "Str0ngP@ssw0rd!",
		"confirmPassword": "Str0ngP@ssw0rd!",
		"birthDate": "1990-01-01T00:00:00Z"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userServiceMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("internal error"))

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"status":500,"title":"Internal Server Error","details":"Something went wrong on our end. Please try again later or contact support if the issue persists."}`, rec.Body.String())
	}
}

func TestUserHandler_Create_WhenSuccessful_ShouldReturnCreated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"firstName": "John",
		"lastName": "Doe",
		"email": "test@example.com",
		"confirmEmail": "test@example.com",
		"password": "Str0ngP@ssw0rd!",
		"confirmPassword": "Str0ngP@ssw0rd!",
		"birthDate": "1990-01-01T00:00:00Z"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userServiceMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "", rec.Body.String()) // Verifica que a resposta não tem corpo (está vazia)
	}
}

func TestUserHandler_Create_WhenBirthDateIsInvalid_ShouldReturnUnprocessableEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
		"firstName": "John",
		"lastName": "Doe",
		"email": "test@example.com",
		"confirmEmail": "test@example.com",
		"password": "Str0ngP@ssw0rd!",
		"confirmPassword": "Str0ngP@ssw0rd!",
		"birthDate": "invalid-date-format"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.Create(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{"status":422,"title":"Unable to Process Request","details":"We encountered an issue while trying to process your request. The data you provided is not in the expected format.","errors":[{"field":"payload","message":"The information provided is not correctly formatted or is missing required fields. Please review and try again."}]}`, rec.Body.String())
	}
}

func TestUserHandler_SignIn_WhenJSONCannotBeDecoded_ShouldReturnUnprocessableEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.SignIn(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{"status":422,"title":"Unable to Process Request","details":"We encountered an issue while trying to process your request. The data you provided is not in the expected format.","errors":[{"field":"payload","message":"The information provided is not correctly formatted or is missing required fields. Please review and try again."}]}`, rec.Body.String())
	}
}

func TestUserHandler_SignIn_WhenValidationFails_ShouldReturnUnprocessableEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{
		"email": "invalid-email",
		"password": ""
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.SignIn(ctx)) {
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.JSONEq(t, `{
			"status": 422,
			"title": "Validation Error",
			"details": "One or more fields are invalid.",
			"errors": [
				{
					"field": "email",
					"message": "Invalid email format"
				},
				{
					"field": "password",
					"message": "This field is required"
				}
			]
		}`, rec.Body.String())
	}
}

func TestUserHandler_SignIn_WhenUserNotFound_ShouldReturnUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBufferString(`{
		"email": "nonexistent@example.com",
		"password": "Str0ngP@ssw0rd!"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userServiceMock.EXPECT().SignIn(gomock.Any(), gomock.Any()).Return(nil, domain.ErrUserNotFound)

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.SignIn(ctx)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.JSONEq(t, `{
			"status": 401,
			"title": "Unauthorized credentials",
			"details": "Unauthorized credentials. Review the data sent."
		}`, rec.Body.String())
	}
}

func TestUserHandler_SignIn_WhenPasswordIsInvalid_ShouldReturnUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{
		"email": "test@example.com",
		"password": "WrongP@ssw0rd"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userServiceMock.EXPECT().SignIn(gomock.Any(), gomock.Any()).Return(nil, domain.ErrInvalidPassword)

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.SignIn(ctx)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.JSONEq(t, `{
			"status": 401,
			"title": "Unauthorized credentials",
			"details": "Unauthorized credentials. Review the data sent."
		}`, rec.Body.String())
	}
}

func TestUserHandler_SignIn_WhenInternalServerErrorOccurs_ShouldReturnInternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBufferString(`{
		"email": "test@example.com",
		"password": "Str0ngP@ssw0rd!"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	userServiceMock.EXPECT().SignIn(gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error"))

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.SignIn(ctx)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"status":500,"title":"Internal Server Error","details":"Something went wrong on our end. Please try again later or contact support if the issue persists."}`, rec.Body.String())
	}
}

func TestUserHandler_SignIn_WhenSuccessful_ShouldReturnOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userServiceMock := mock.NewMockUserService(ctrl)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{
		"email": "test@example.com",
		"password": "Str0ngP@ssw0rd!"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetRequest(req.WithContext(context.Background()))

	expectedResponse := &domain.SignInResponse{
		Token: "valid-token",
	}
	userServiceMock.EXPECT().SignIn(gomock.Any(), gomock.Any()).Return(expectedResponse, nil)

	userHandler := &userHandler{
		userService: userServiceMock,
	}

	if assert.NoError(t, userHandler.SignIn(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"token":"valid-token"}`, rec.Body.String())
	}
}
