package model

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
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
	GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_INVALIED = 0
	GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_AND_WRITE = 1
	GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_ONLY      = 2
)

func NewGoogleSheetsOAuth2Claims() *GoogleSheetsOAuth2Claims {
	return &GoogleSheetsOAuth2Claims{}
}

func GenerateGoogleSheetsOAuth2Token(teamID, userID, resourceID, createOAuthTokenRequest *request.CreateOAuthTokenRequest) (string, error) {
	accessType : =  GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_ONLY
	if createOAuthTokenRequest.IsReadAndWrite(){
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
				Time: time.Now().Add(time.Minute * 1),
			},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	conf := config.GetInstance()
	accessToken, err := token.SignedString([]byte(conf.GetSecretKey()))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}


func (i *GoogleSheetsOAuth2Claims) ValidateAccessToken(accessToken stirng) (int, error) {
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
