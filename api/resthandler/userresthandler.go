// Copyright 2022 The ILLA Authors.

//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resthandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/illa-family/builder-backend/pkg/user"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Email struct {
	Email string `json:"email" validate:"required"`
}

type Username struct {
	Username string `json:"username" validate:"required"`
}

type Language struct {
	Language string `json:"language" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required"`
}

type SignUpRequest struct {
	Username          string `json:"username" validate:"required"`
	Email             string `json:"email" validate:"required"`
	Password          string `json:"password" validate:"required"`
	Language          string `json:"language" validate:"required"`
	IsSubscribed      bool   `json:"isSubscribed"`
	VerificationCode  string `json:"verificationCode" validate:"required"`
	VerificationToken string `json:"verificationToken" validate:"required"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type ForgetPasswordRequest struct {
	Email             string `json:"email" validate:"required"`
	NewPassword       string `json:"newPassword" validate:"required"`
	VerificationCode  string `json:"verificationCode" validate:"required"`
	VerificationToken string `json:"verificationToken" validate:"required"`
}

type UserRestHandler interface {
	GetVerificationCode(c *gin.Context)
	SignUp(c *gin.Context)
	SignIn(c *gin.Context)
	ForgetPassword(c *gin.Context)
	UpdateUsername(c *gin.Context)
	UpdatePassword(c *gin.Context)
	UpdateLanguage(c *gin.Context)
}

type UserRestHandlerImpl struct {
	logger      *zap.SugaredLogger
	userService user.UserService
}

func NewUserRestHandlerImpl(logger *zap.SugaredLogger, userService user.UserService) *UserRestHandlerImpl {
	return &UserRestHandlerImpl{
		logger:      logger,
		userService: userService,
	}
}

func (impl UserRestHandlerImpl) GetVerificationCode(c *gin.Context) {
	var payload Email
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	vToken, err := impl.userService.GenerateVerificationCode(payload.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"verificationToken": vToken,
	})
}

func (impl UserRestHandlerImpl) SignUp(c *gin.Context) {
	// get request body
	var payload SignUpRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// eliminate duplicate user
	if duplicateUser, _ := impl.userService.FindUserByEmail(payload.Email); duplicateUser.UserId != uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": errors.New("duplicate email address").Error(),
		})
		return
	}

	// validate verification code
	validCode, err := impl.userService.ValidateVerificationCode(payload.VerificationCode, payload.VerificationToken)
	if !validCode || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": errors.New("invalid verification code").Error(),
		})
		return
	}

	// create user
	userDto, err := impl.userService.CreateUser(&user.UserDto{
		Username:     payload.Username,
		Password:     payload.Password,
		Email:        payload.Email,
		Language:     payload.Language,
		IsSubscribed: payload.IsSubscribed,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": errors.New("sign up error").Error(),
		})
		return
	}

	// generate access token and refresh token
	accessToken, refreshToken, _ := impl.userService.GetToken(userDto.UserId)
	c.SetCookie("access_token", accessToken, 7200, "/", "localhost", false, true)
	c.SetCookie("refresh_token", refreshToken, 259200, "/", "localhost", false, true)

	c.JSON(http.StatusOK, userDto)
}

func (impl UserRestHandlerImpl) SignIn(c *gin.Context) {
	// get request body
	var payload SignInRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// fetch user by email
	userDto, err := impl.userService.FindUserByEmail(payload.Email)
	if err != nil || userDto.UserId == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": errors.New("no such user").Error(),
		})
		return
	}

	// validate password with password digest
	err = bcrypt.CompareHashAndPassword([]byte(userDto.Password), []byte(payload.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": errors.New("invalid password").Error(),
		})
		return
	}

	// generate access token and refresh token
	accessToken, refreshToken, _ := impl.userService.GetToken(userDto.UserId)
	c.SetCookie("access_token", accessToken, 7200, "/", "localhost", false, true)
	c.SetCookie("refresh_token", refreshToken, 259200, "/", "localhost", false, true)

	c.JSON(http.StatusOK, userDto)
}

func (impl UserRestHandlerImpl) ForgetPassword(c *gin.Context) {
	// get request body
	var payload ForgetPasswordRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// fetch user by email
	userDto, err := impl.userService.FindUserByEmail(payload.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	if userDto.UserId == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": errors.New("no such user, please sign up").Error(),
		})
		return
	}

	// validate verification code
	validCode, err := impl.userService.ValidateVerificationCode(payload.VerificationCode, payload.VerificationToken)
	if !validCode || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": errors.New("invalid verification code").Error(),
		})
		return
	}

	// update password
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	userDto.Password = string(hashPwd)
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "change password successfully",
	})
}

func (impl UserRestHandlerImpl) UpdateUsername(c *gin.Context) {
	// get request body
	var payload Username
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// get user by id
	userID, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": errors.New("unauthorized").Error(),
		})
		return
	}
	userId, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": errors.New("unauthorized").Error(),
		})
		return
	}
	id, err := uuid.Parse(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	userDto, err := impl.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}

	// update user username
	userDto.Language = payload.Username
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userDto)

}

func (impl UserRestHandlerImpl) UpdatePassword(c *gin.Context) {
	// get request body
	var payload ChangePasswordRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// get user by id
	userID, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": errors.New("unauthorized").Error(),
		})
		return
	}
	userId, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": errors.New("unauthorized").Error(),
		})
		return
	}
	id, err := uuid.Parse(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	userDto, err := impl.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate current password with password digest
	if err := bcrypt.CompareHashAndPassword([]byte(userDto.Password), []byte(payload.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// update password
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	userDto.Password = string(hashPwd)
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userDto)
}

func (impl UserRestHandlerImpl) UpdateLanguage(c *gin.Context) {
	// get request body
	var payload Language
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}

	// get user by id
	userID, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": errors.New("unauthorized").Error(),
		})
		return
	}
	userId, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": errors.New("unauthorized").Error(),
		})
		return
	}
	id, err := uuid.Parse(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}
	userDto, err := impl.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}

	// update user language
	userDto.Language = payload.Language
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorCode":    500,
			"errorMessage": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userDto)
}
