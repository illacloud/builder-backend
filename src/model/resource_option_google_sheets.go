package model

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const (
	GOOGLE_SHEET_OAUTH_TYPE = "oauth2"
)

type GoogleSheetsOAuth2Options struct {
	AccessType   string `json:"accessType"`
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType"`
	RefreshToken string `json:"refreshToken"`
	Status       int    `json:"status"`
}

type ResourceOptionGoogleSheets struct {
	Authentication string                     `json:"authentication"`
	Options        *GoogleSheetsOAuth2Options `json:"opts"`
}

func NewResourceOptionGoogleSheetsByResource(resource *Resource) (*ResourceOptionGoogleSheets, error) {
	fmt.Printf("[DUMP] NewResourceOptionGoogleSheetsByResource().resource: %+v\n", NewResourceOptionGoogleSheetsByResource)
	fmt.Printf("[DUMP] NewResourceOptionGoogleSheetsByResource().resource.Options: %+v\n", resource.Options)
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
