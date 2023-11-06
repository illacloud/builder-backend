package illadrivesdk

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/illacloud/builder-backend/src/utils/config"
)

type DriveAuthClaims struct {
	TeamID  int    `json:"teamID"`
	DriveID int    `json:"driveID"`
	Usage   string `json:"usage"`
	jwt.RegisteredClaims
}

const JWT_ISSUER = "ILLA Cloud"
const JWT_TOKEN_DEFAULT_EXIPRED_PERIOD = time.Hour * 24

func GenerateAndSendVerificationCode(teamID int, driveID int, usage string) (string, error) {
	claims := &DriveAuthClaims{
		TeamID:  teamID,
		DriveID: driveID,
		Usage:   usage,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: JWT_ISSUER,
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(JWT_TOKEN_DEFAULT_EXIPRED_PERIOD),
			},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	conf := config.GetInstance()
	codeToken, err := token.SignedString([]byte(conf.GetSecretKey()))
	if err != nil {
		return "", err
	}

	return codeToken, nil
}

func Validate(jwtToken string, teamID int, driveID int, usage string) (bool, error) {
	// parse token for start with "bearer"
	jwtTokenFinal := jwtToken
	jwtTokenSplited := strings.Split(jwtToken, " ")
	if len(jwtTokenSplited) == 2 {
		jwtTokenFinal = jwtTokenSplited[1]
	}
	// check
	defaultClaims := &DriveAuthClaims{}
	token, err := jwt.ParseWithClaims(jwtTokenFinal, defaultClaims, func(token *jwt.Token) (interface{}, error) {
		conf := config.GetInstance()
		return []byte(conf.GetSecretKey()), nil
	})
	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(*DriveAuthClaims)
	if !(claims.TeamID == teamID) {
		return false, errors.New("invalied team ID")
	}
	if !(claims.DriveID == driveID) {
		return false, errors.New("invalied drive ID")
	}
	if !(ok && claims.Usage == usage) {
		return false, errors.New("invalied token usage")
	}

	return true, nil
}
