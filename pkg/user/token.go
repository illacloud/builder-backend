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
)

func CreateAccessToken(id uuid.UUID) (string, error) {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vCode := fmt.Sprintf("%06v", rnd.Int31n(10000))

	claims := jwt.MapClaims{}
	claims["userId"] = id.String()
	claims["rnd"] = vCode
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(os.Getenv("ILLA_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func CreateRefreshToken(accessToken string) (string, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vCode := fmt.Sprintf("%06v", rnd.Int31n(10000))

	claims := jwt.MapClaims{}
	claims["token"] = accessToken
	claims["rnd"] = vCode
	claims["exp"] = time.Now().Add(time.Hour * 24 * 3).Unix()

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("ILLA_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return refreshTokenString, err
}

func ValidateAccessToken(accessToken string) (bool, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("ILLA_SECRET_KEY")), nil
	})
	if err != nil {
		return false, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		return false, err
	}
	userId := payload["userId"].(string)
	if _, err := uuid.Parse(userId); err != nil {
		return false, err
	}

	return true, nil
}

func ValidateRefreshToken(refreshToken string) (bool, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("ILLA_SECRET_KEY")), nil
	})
	if err != nil {
		return false, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		return false, errors.New("invalid token")
	}

	_, ok = payload["token"].(string)
	if !ok {
		return false, errors.New("invalid token")
	}

	return true, nil
}

func ExtractUserIdFromToken(accessToken string) (uuid.UUID, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("ILLA_SECRET_KEY")), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, err
	}
	userId := payload["userId"].(string)
	userID, err := uuid.Parse(userId)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}
