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
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/repository"
	"go.uber.org/zap"
)

type AuthClaims struct {
	User   int       `json:"user"`
	UUID   uuid.UUID `json:"uuid"`
	Random string    `json:"rnd"`
	jwt.RegisteredClaims
}

type Authenticator interface {
	ValidateAccessToken(accessToken string) (bool, error)
	ExtractUserIDFromToken(accessToken string) (int, uuid.UUID, error)
	ValidateUser(id int, uid uuid.UUID) (bool, error)
	ValidateUserAndGetDetail(id int, uid uuid.UUID) (bool, *repository.User, error)
}

type AuthenticatorImpl struct {
	logger         *zap.SugaredLogger
	userRepository repository.UserRepository
}

func NewAuthenticatorImpl(userRepository repository.UserRepository, logger *zap.SugaredLogger) *AuthenticatorImpl {
	return &AuthenticatorImpl{
		logger:         logger,
		userRepository: userRepository,
	}
}

func (impl *AuthenticatorImpl) ValidateAccessToken(accessToken string) (bool, error) {
	_, _, err := impl.ExtractUserIDFromToken(accessToken)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl *AuthenticatorImpl) ExtractUserIDFromToken(accessToken string) (int, uuid.UUID, error) {
	authClaims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(accessToken, authClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ILLA_SECRET_KEY")), nil
	})
	if err != nil {
		return 0, uuid.Nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !(ok && token.Valid) {
		return 0, uuid.Nil, err
	}

	return claims.User, claims.UUID, nil
}

func (impl *AuthenticatorImpl) ValidateUser(id int, uid uuid.UUID) (bool, error) {
	userRecord, err := impl.userRepository.FetchUserByUKey(id, uid)
	if err != nil {
		return false, err
	}
	if userRecord.ID != id || userRecord.UID != uid {
		return false, errors.New("no such user")
	}

	return true, nil
}

func (impl *AuthenticatorImpl) ValidateUserAndGetDetail(id int, uid uuid.UUID) (bool, *repository.User, error) {
	userRecord, err := impl.userRepository.FetchUserByUKey(id, uid)
	if err != nil {
		return false, nil, err
	}
	if userRecord.ID != id || userRecord.UID != uid {
		return false, nil, errors.New("no such user")
	}
	return true, userRecord, nil
}

func CreateAccessToken(id int, uid uuid.UUID) (string, error) {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vCode := fmt.Sprintf("%06v", rnd.Int31n(10000))

	claims := &AuthClaims{
		User:   id,
		UUID:   uid,
		Random: vCode,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "ILLA",
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(time.Hour * 24 * 7),
			},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(os.Getenv("ILLA_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
