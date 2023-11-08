package illadrivesdk

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/illacloud/builder-backend/src/utils/config"
)

const (
	DRIVE_API_ACTION_LIST                        = "List"
	DRIVE_API_ACTION_READ_FILE_PROPERTY          = "ReadFileProperty"
	DRIVE_API_ACTION_GET_UPLOAD_ADDRES           = "GetUploadAddres"
	DRIVE_API_ACTION_GET_MUTIPLE_UPLOAD_ADDRES   = "GetMutipleUploadAddres"
	DRIVE_API_ACTION_GET_DOWNLOAD_ADDRES         = "GetDownloadAddres"
	DRIVE_API_ACTION_GET_MUTIPLE_DOWNLOAD_ADDRES = "GetMutipleDownloadAddres"
	DRIVE_API_ACTION_DELETE_FILE                 = "DeleteFile"
	DRIVE_API_ACTION_DELETE_MULTIPLE_FILE        = "DeleteMultipleFile"
	DRIVE_API_ACTION_UPDATE_FILE_PROPERTY        = "UpdateFileProperty"
	DRIVE_API_ACTION_GENERATE_TINY_URLS          = "GenerateTinyURLs"
	DRIVE_API_ACTION_GET_DOLDER_ID_BY_PATH       = "GetFolderIDByPath"
)

type DriveAuthClaims struct {
	TeamID int    `json:"teamID"`
	Action string `json:"action"`
	jwt.RegisteredClaims
}

const JWT_ISSUER = "ILLA Cloud"
const JWT_TOKEN_DEFAULT_EXIPRED_PERIOD = time.Hour * 24

func GenerateDriveAPIActionToken(teamID int, action string) (string, error) {
	claims := &DriveAuthClaims{
		TeamID: teamID,
		Action: action,
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

func Validate(jwtToken string, teamID int, action string) (bool, error) {
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
	if !(ok && claims.Action == action) {
		return false, errors.New("invalied token action")
	}

	return true, nil
}
