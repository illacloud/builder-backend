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

package user

import (
	"time"

	"github.com/illa-family/builder-backend/pkg/smtp"
	"github.com/illa-family/builder-backend/pkg/user/repository"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(userDto *UserDto) (*UserDto, error)
	UpdateUser(userDto *UserDto) (*UserDto, error)
	FindUserByEmail(email string) (*UserDto, error)
	GetUser(userId uuid.UUID) (*UserDto, error)
	GenerateVerificationCode(email, usage string) (string, error)
	ValidateVerificationCode(vCode, vToken, usage string) (bool, error)
	GetToken(userId uuid.UUID) (string, error)
}

type UserDto struct {
	UserId       uuid.UUID `json:"userId,omitempty"`
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"-"`
	Email        string    `json:"email,omitempty"`
	Language     string    `json:"language,omitempty"`
	IsSubscribed bool      `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type UserServiceImpl struct {
	logger         *zap.SugaredLogger
	userRepository repository.UserRepository
	smtpServer     smtp.SMTPServer
}

func NewUserServiceImpl(userRepository repository.UserRepository, logger *zap.SugaredLogger,
	smtpServer smtp.SMTPServer) *UserServiceImpl {
	return &UserServiceImpl{
		logger:         logger,
		userRepository: userRepository,
		smtpServer:     smtpServer,
	}
}

func (impl *UserServiceImpl) CreateUser(userDto *UserDto) (*UserDto, error) {
	userDto.UserId = uuid.New()
	userDto.CreatedAt = time.Now().UTC()
	userDto.UpdatedAt = time.Now().UTC()
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(userDto.Password), bcrypt.DefaultCost)
	if err != nil {
		return &UserDto{}, err
	}
	if err := impl.userRepository.CreateUser(&repository.User{
		ID:             userDto.UserId,
		Username:       userDto.Username,
		PasswordDigest: string(hashPwd),
		Email:          userDto.Email,
		Language:       userDto.Language,
		IsSubscribed:   userDto.IsSubscribed,
		CreatedAt:      userDto.CreatedAt,
		UpdatedAt:      userDto.UpdatedAt,
	}); err != nil {
		return &UserDto{}, err
	}
	return userDto, nil
}

func (impl *UserServiceImpl) UpdateUser(userDto *UserDto) (*UserDto, error) {
	userDto.UpdatedAt = time.Now().UTC()
	if err := impl.userRepository.UpdateUser(&repository.User{
		ID:             userDto.UserId,
		Username:       userDto.Username,
		PasswordDigest: userDto.Password,
		Email:          userDto.Email,
		Language:       userDto.Language,
		IsSubscribed:   userDto.IsSubscribed,
		CreatedAt:      userDto.CreatedAt,
		UpdatedAt:      userDto.UpdatedAt,
	}); err != nil {
		return &UserDto{}, err
	}
	return userDto, nil
}

func (impl *UserServiceImpl) FindUserByEmail(email string) (*UserDto, error) {
	userRecord, err := impl.userRepository.FetchUserByEmail(email)
	if err != nil {
		return &UserDto{}, err
	}
	userDto := &UserDto{
		UserId:       userRecord.ID,
		Username:     userRecord.Username,
		Email:        userRecord.Email,
		Password:     userRecord.PasswordDigest,
		Language:     userRecord.Language,
		IsSubscribed: userRecord.IsSubscribed,
	}
	return userDto, nil
}

func (impl *UserServiceImpl) GenerateVerificationCode(email, usage string) (string, error) {
	return impl.smtpServer.NewVerificationCode(email, usage)
}

func (impl *UserServiceImpl) ValidateVerificationCode(vCode, vToken, usage string) (bool, error) {
	return impl.smtpServer.ValidateVerificationCode(vToken, vCode, usage)
}

func (impl *UserServiceImpl) GetToken(userId uuid.UUID) (string, error) {
	accessToken, err := CreateAccessToken(userId)
	if err != nil {
		return "", nil
	}
	return accessToken, nil
}

func (impl *UserServiceImpl) GetUser(userId uuid.UUID) (*UserDto, error) {
	userRecord, err := impl.userRepository.RetrieveById(userId)
	if err != nil {
		return &UserDto{}, err
	}
	userDto := &UserDto{
		UserId:       userRecord.ID,
		Username:     userRecord.Username,
		Email:        userRecord.Email,
		Password:     userRecord.PasswordDigest,
		Language:     userRecord.Language,
		IsSubscribed: userRecord.IsSubscribed,
		CreatedAt:    userRecord.CreatedAt,
		UpdatedAt:    userRecord.UpdatedAt,
	}
	return userDto, nil
}
