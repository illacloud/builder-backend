package jwtauthenticator

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const JWT_ISSUER = "ILLA Cloud"
const JWT_TOKEN_DEFAULT_EXIPRED_PERIOD = time.Minute * 30

type DefaultClaims struct {
	Username string `json:"username"`
	Usage    string `json:"usage"`
	jwt.RegisteredClaims
}

func GenerateAndSendVerificationCode(username, usage string) (string, error) {
	claims := &DefaultClaims{
		Username: username,
		Usage:    usage,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: JWT_ISSUER,
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(JWT_TOKEN_DEFAULT_EXIPRED_PERIOD),
			},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	codeToken, err := token.SignedString([]byte(os.Getenv("ILLA_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return codeToken, nil
}

func Validate(jwtToken, username, usage string) (bool, error) {
	// parse token for start with "bearer"
	jwtTokenFinal := jwtToken
	jwtTokenSplited := strings.Split(jwtToken, " ")
	if len(jwtTokenSplited) == 2 {
		jwtTokenFinal = jwtTokenSplited[1]
	}
	// check
	defaultClaims := &DefaultClaims{}
	token, err := jwt.ParseWithClaims(jwtTokenFinal, defaultClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ILLA_SECRET_KEY")), nil
	})
	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(*DefaultClaims)
	if !(ok && claims.Usage == usage) {
		return false, errors.New("invalied token usage")
	}
	if !(claims.Username == username) {
		return false, errors.New("invalied token")
	}
	return true, nil
}
