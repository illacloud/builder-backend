package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/utils/config"
)

type GoogleSheetsOAuth2Claims struct {
	Team     int    `json:"team"`
	User     int    `json:"user"`
	Resource int    `json:"resource"`
	Access   int    `json:"access"`
	URL      string `json:"url"`
	jwt.RegisteredClaims
}

const (
	GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_INVALIED       = 0
	GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_AND_WRITE = 1
	GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_ONLY      = 2
)

const (
	GOOGLE_SHEETS_OAUTH_STATUS_SUCCESS = 1
	GOOGLE_SHEETS_OAUTH_STATUS_FAILED  = 2
)

const (
	GOOGLE_OAUTH2_TOKEN_DEFAULT_EXIPRED_PERIOD = time.Minute * 60
)

func NewGoogleSheetsOAuth2Claims() *GoogleSheetsOAuth2Claims {
	return &GoogleSheetsOAuth2Claims{}
}

func GenerateGoogleSheetsOAuth2Token(teamID int, userID int, resourceID int, createOAuthTokenRequest *request.CreateOAuthTokenRequest) (string, error) {
	fmt.Printf("[DUMP] GenerateGoogleSheetsOAuth2Token.createOAuthTokenRequest: %+v\n", createOAuthTokenRequest)
	fmt.Printf("[DUMP] createOAuthTokenRequest.IsReadAndWrite(): %+v\n", createOAuthTokenRequest.IsReadAndWrite())
	accessType := GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_ONLY
	if createOAuthTokenRequest.IsReadAndWrite() {
		accessType = GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_AND_WRITE
	}
	claims := &GoogleSheetsOAuth2Claims{
		Team:     teamID,
		User:     userID,
		Resource: resourceID,
		Access:   accessType,
		URL:      createOAuthTokenRequest.ExportRedirectURL(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "ILLA",
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(GOOGLE_OAUTH2_TOKEN_DEFAULT_EXIPRED_PERIOD),
			},
		},
	}
	fmt.Printf("[DUMP] createOAuthTokenRequest.claims(): %+v\n", claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	conf := config.GetInstance()
	accessToken, err := token.SignedString([]byte(conf.GetSecretKey()))
	fmt.Printf("[DUMP] createOAuthTokenRequest.token(): %+v\n", token)
	fmt.Printf("[DUMP] createOAuthTokenRequest.accessToken(): %+v\n", accessToken)

	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (i *GoogleSheetsOAuth2Claims) ExtractGoogleSheetsOAuth2TokenInfo(accessToken string) (teamID, userID, resourceID int, url string, err error) {
	token, errInParseClaims := jwt.ParseWithClaims(accessToken, i, func(token *jwt.Token) (interface{}, error) {
		conf := config.GetInstance()
		return []byte(conf.GetSecretKey()), nil
	})
	if errInParseClaims != nil {
		return 0, 0, 0, "", errInParseClaims
	}

	claims, assertPass := token.Claims.(*GoogleSheetsOAuth2Claims)
	if !(assertPass && token.Valid) {
		return 0, 0, 0, "", errors.New("invalied access token")
	}

	return claims.Team, claims.User, claims.Resource, claims.URL, nil
}

func (i *GoogleSheetsOAuth2Claims) ValidateAccessToken(accessToken string) (int, error) {
	token, errInParseClaims := jwt.ParseWithClaims(accessToken, i, func(token *jwt.Token) (interface{}, error) {
		conf := config.GetInstance()
		return []byte(conf.GetSecretKey()), nil
	})
	if errInParseClaims != nil {
		return GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_INVALIED, errInParseClaims
	}

	claims, assertPass := token.Claims.(*GoogleSheetsOAuth2Claims)
	if !(assertPass && token.Valid) {
		return GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_INVALIED, errors.New("invalied access token")
	}

	accessType := claims.Access
	return accessType, nil
}
