package model

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

const (
	GOOGLE_SHEET_OAUTH_TYPE = "oauth2"
)

type GoogleSheetsOAuth2Options struct {
	AccessType   string
	AccessToken  string
	TokenType    string
	RefreshToken string
	Status       int
}

type ResourceOptionGoogleSheets struct {
	Authentication string
	Options        *GoogleSheetsOAuth2Options
}

func NewResourceOptionGoogleSheetsByResource(resource *Resource) (*ResourceOptionGoogleSheets, error) {
	resourceOptionGoogleSheets := &ResourceOptionGoogleSheets{}
	errInDecode := mapstructure.Decode(resource.Options, &resourceOptionGoogleSheets)
	if errInDecode != nil {
		return nil, errInDecode
	}
	return resourceOptionGoogleSheets, nil
}

func (i *ResourceOptionGoogleSheets) IsAvaliableAuthenticationMethod() bool {
	return i.Authentication == GOOGLE_SHEET_OAUTH_TYPE
}

func (i *ResourceOptionGoogleSheets) SetAccessToken(accessToken string) {
	i.Options.AccessToken = accessToken
}

func (i *ResourceOptionGoogleSheets) ExportRefreshToken() string {
	return i.Options.RefreshToken
}

func (i *ResourceOptionGoogleSheets) ExportInString() string {
	byteData, _ := json.Marshal(i)
	return string(byteData)
}
