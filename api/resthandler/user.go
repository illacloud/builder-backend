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
	"net/http"
	"os"
	"time"

	"github.com/illacloud/builder-backend/pkg/user"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type VerificationRequest struct {
	Email string `json:"email" validate:"required"`
	Usage string `json:"usage" validate:"oneof=signup forgetpwd"`
}

type Username struct {
	Nickname string `json:"nickname" validate:"required"`
}

type Language struct {
	Language string `json:"language" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required"`
}

type SignUpRequest struct {
	Nickname          string `json:"nickname" validate:"required"`
	Email             string `json:"email" validate:"required"`
	Password          string `json:"password" validate:"required"`
	Language          string `json:"language" validate:"oneof=zh-CN en-US ko-KR ja-JP"`
	IsSubscribed      bool   `json:"isSubscribed"`
	VerificationCode  string `json:"verificationCode"`
	VerificationToken string `json:"verificationToken"`
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
	GetUserInfo(c *gin.Context)
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
	var payload VerificationRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	vToken, err := impl.userService.GenerateVerificationCode(payload.Email, payload.Usage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "send verification code error: " + err.Error(),
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
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// eliminate duplicate user
	if duplicateUser, _ := impl.userService.FindUserByEmail(payload.Email); duplicateUser.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "duplicate email address",
		})
		return
	}

	// validate verification code
	if os.Getenv("ILLA_DEPLOY_MODE") == "cloud" {
		validCode, err := impl.userService.ValidateVerificationCode(payload.VerificationCode, payload.VerificationToken,
			payload.Email, "signup")
		if err != nil || !validCode {
			c.JSON(http.StatusBadRequest, gin.H{
				"errorCode":    400,
				"errorMessage": "validate verification code error: " + err.Error(),
			})
			return
		}
	}

	// create user
	userDto, err := impl.userService.CreateUser(user.UserDto{
		Nickname:     payload.Nickname,
		Password:     payload.Password,
		Email:        payload.Email,
		Language:     payload.Language,
		IsSubscribed: payload.IsSubscribed,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "sign up error: " + err.Error(),
		})
		return
	}

	if payload.IsSubscribed {
		_ = impl.userService.SendSubscriptionEmail(payload.Email)
	}

	// generate access token and refresh token
	accessToken, _ := impl.userService.GetToken(userDto.ID, userDto.UID)
	c.Header("illa-token", accessToken)

	c.JSON(http.StatusOK, userDto)
}

func (impl UserRestHandlerImpl) SignIn(c *gin.Context) {
	// get request body
	var payload SignInRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// fetch user by email
	userDto, err := impl.userService.FindUserByEmail(payload.Email)
	if err != nil || userDto.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "invalid email or password",
		})
		return
	}

	// validate password with password digest
	err = bcrypt.CompareHashAndPassword([]byte(userDto.Password), []byte(payload.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "invalid email or password",
		})
		return
	}

	// generate access token and refresh token
	accessToken, _ := impl.userService.GetToken(userDto.ID, userDto.UID)
	c.Header("illa-token", accessToken)

	c.JSON(http.StatusOK, userDto)
}

func (impl UserRestHandlerImpl) ForgetPassword(c *gin.Context) {
	// get request body
	var payload ForgetPasswordRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// fetch user by email
	userDto, err := impl.userService.FindUserByEmail(payload.Email)
	if err != nil || userDto.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "no such user",
		})
		return
	}

	// validate verification code
	validCode, err := impl.userService.ValidateVerificationCode(payload.VerificationCode, payload.VerificationToken,
		payload.Email, "forgetpwd")
	if !validCode || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "validate verification code error: " + err.Error(),
		})
		return
	}

	// update password
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "forget password error: " + err.Error(),
		})
		return
	}
	userDto.Password = string(hashPwd)
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "forget password error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "reset password successfully",
	})
}

func (impl UserRestHandlerImpl) UpdateUsername(c *gin.Context) {
	// get request body
	var payload Username
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// get user by id
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}
	userDto, err := impl.userService.GetUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get user error: " + err.Error(),
		})
		return
	}

	// update user Nickname
	userDto.Nickname = payload.Nickname
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "update user error: " + err.Error(),
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
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// get user by id
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}
	userDto, err := impl.userService.GetUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get user error: " + err.Error(),
		})
		return
	}

	// validate current password with password digest
	if err := bcrypt.CompareHashAndPassword([]byte(userDto.Password), []byte(payload.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "current password incorrect",
		})
		return
	}

	// update password
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "update password error: " + err.Error(),
		})
		return
	}
	userDto.Password = string(hashPwd)
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "update password error: " + err.Error(),
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
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// validate payload required fields
	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "parse request body error: " + err.Error(),
		})
		return
	}

	// get user by id
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}
	userDto, err := impl.userService.GetUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "get user error: " + err.Error(),
		})
		return
	}

	// update user language
	userDto.Language = payload.Language
	if _, err := impl.userService.UpdateUser(userDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": "update language error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userDto)
}

func (impl UserRestHandlerImpl) GetUserInfo(c *gin.Context) {
	// get user by id
	userID, okGet := c.Get("userID")
	user, okReflect := userID.(int)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}
	userDto, err := impl.userService.GetUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode":    400,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, userDto)
}
