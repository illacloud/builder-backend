package illadrivesdk

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/illacloud/builder-backend/src/utils/config"
)

const (
	DRIVE_API_ACTION_GENERATE_TINY_URLS            = "GenerateTinyURLs"
	DRIVE_API_ACTION_GET_FOLDER_ID_BY_PATH         = "GetFolderIDByPath"
	DRIVE_API_ACTION_LIST_FILES                    = "ListFiles"
	DRIVE_API_ACTION_GET_UPLOAD_ADDRESS            = "GetUploadAddress"
	DRIVE_API_ACTION_UPDATE_FILE_STATUS            = "UpdateFileStatus"
	DRIVE_API_ACTION_GET_MULTIPLE_UPLOAD_ADDRESS   = "GetMultipleUploadAddress"
	DRIVE_API_ACTION_GET_DOWNLOAD_ADDRESS          = "GetDownloadAddress"
	DRIVE_API_ACTION_GET_MULTIPLE_DOWNLOAD_ADDRESS = "GetMultipleDownloadAddress"
	DRIVE_API_ACTION_DELETE_FILES                  = "DeleteFiles"
	DRIVE_API_ACTION_DELETE_FILE                   = "DeleteFile"
	DRIVE_API_ACTION_DELETE_MULTIPLE_FILE          = "DeleteMultipleFile"
	DRIVE_API_ACTION_RENAME_FILE                   = "RenameFile"
)

type DriveAuthClaims struct {
	TeamID       int    `json:"teamID"`
	Action       string `json:"action"`
	UserID       int    `json:"userID"`
	InstanceType int    `json:"instanceType"`
	InstanceID   int    `json:"instanceID"`
	jwt.RegisteredClaims
}

const JWT_ISSUER = "ILLA Cloud"
const JWT_TOKEN_DEFAULT_EXIPRED_PERIOD = time.Hour * 24

func GenerateDriveAPIActionToken(api *IllaDriveRestAPI, action string) (string, error) {
	claims := &DriveAuthClaims{
		TeamID:       api.TeamID,
		Action:       action,
		UserID:       api.UserID,
		InstanceType: api.InstanceType,
		InstanceID:   api.InstanceID,
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
