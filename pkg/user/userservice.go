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
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/smtp"
	"golang.org/x/crypto/bcrypt"

	"go.uber.org/zap"
)

var language_array = []string{"", "en-US", "zh-CN", "ko-KR", "ja-JP"}
var language_map = map[string]int{
	"en-US": 1,
	"zh-CN": 2,
	"ko-KR": 3,
	"ja-JP": 4,
}

type UserService interface {
	CreateUser(userDto UserDto) (UserDto, error)
	UpdateUser(userDto UserDto) (UserDto, error)
	FindUserByEmail(email string) (UserDto, error)
	GetUser(id int) (UserDto, error)
	GetToken(id int, uid uuid.UUID) (string, error)
	GenerateVerificationCode(email, usage string) (string, error)
	ValidateVerificationCode(vCode, vToken, email, usage string) (bool, error)
	SendSubscriptionEmail(email string) error
}

type UserDto struct {
	ID           int       `json:"-"`
	SID          string    `json:"userId,omitempty"`
	UID          uuid.UUID `json:"-"`
	Nickname     string    `json:"nickname,omitempty"`
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

func (impl *UserServiceImpl) CreateUser(userDto UserDto) (UserDto, error) {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(userDto.Password), bcrypt.DefaultCost)
	if err != nil {
		return UserDto{}, err
	}
	uid := uuid.New()
	id, err := impl.userRepository.CreateUser(&repository.User{
		UID:            uid,
		Nickname:       userDto.Nickname,
		PasswordDigest: string(hashPwd),
		Email:          userDto.Email,
		Language:       language_map[userDto.Language],
		IsSubscribed:   userDto.IsSubscribed,
		CreatedAt:      userDto.CreatedAt,
		UpdatedAt:      userDto.UpdatedAt,
	})
	if err != nil {
		return UserDto{}, err
	}

	userDto.ID = id
	userDto.UID = uid
	userDto.SID = strconv.Itoa(id)

	return userDto, nil
}

func (impl *UserServiceImpl) UpdateUser(userDto UserDto) (UserDto, error) {
	if err := impl.userRepository.UpdateUser(&repository.User{
		ID:             userDto.ID,
		Nickname:       userDto.Nickname,
		PasswordDigest: userDto.Password,
		Language:       language_map[userDto.Language],
		UpdatedAt:      userDto.UpdatedAt,
	}); err != nil {
		return UserDto{}, err
	}
	userDto.SID = strconv.Itoa(userDto.ID)

	return userDto, nil
}

func (impl *UserServiceImpl) FindUserByEmail(email string) (UserDto, error) {
	userRecord, err := impl.userRepository.FetchUserByEmail(email)
	if err != nil {
		return UserDto{}, err
	}
	userDto := UserDto{
		ID:           userRecord.ID,
		SID:          strconv.Itoa(userRecord.ID),
		UID:          userRecord.UID,
		Nickname:     userRecord.Nickname,
		Email:        userRecord.Email,
		Password:     userRecord.PasswordDigest,
		Language:     language_array[userRecord.Language],
		IsSubscribed: userRecord.IsSubscribed,
	}
	return userDto, nil
}

func (impl *UserServiceImpl) GetUser(id int) (UserDto, error) {
	userRecord, err := impl.userRepository.RetrieveByID(id)
	if err != nil {
		return UserDto{}, err
	}
	userDto := UserDto{
		ID:           userRecord.ID,
		SID:          strconv.Itoa(userRecord.ID),
		Nickname:     userRecord.Nickname,
		Email:        userRecord.Email,
		Password:     userRecord.PasswordDigest,
		Language:     language_array[userRecord.Language],
		IsSubscribed: userRecord.IsSubscribed,
		CreatedAt:    userRecord.CreatedAt,
		UpdatedAt:    userRecord.UpdatedAt,
	}
	return userDto, nil
}

func (impl *UserServiceImpl) GetToken(id int, uid uuid.UUID) (string, error) {
	accessToken, err := CreateAccessToken(id, uid)
	if err != nil {
		return "", nil
	}
	return accessToken, nil
}

func (impl *UserServiceImpl) GenerateVerificationCode(email, usage string) (string, error) {
	return impl.smtpServer.NewVerificationCode(email, usage)
}

func (impl *UserServiceImpl) ValidateVerificationCode(vCode, vToken, email, usage string) (bool, error) {
	return impl.smtpServer.ValidateVerificationCode(vToken, vCode, email, usage)
}

func (impl *UserServiceImpl) SendSubscriptionEmail(email string) error {
	return impl.smtpServer.SendSubscriptionEmail(email)
}
