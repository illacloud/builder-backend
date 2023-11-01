package model

import (
	"encoding/json"
	"fmt"

	"github.com/illacloud/builder-backend/src/utils/oauthgoogle"
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
	ExpiresIn    int    `json:"expiresIn"`
	Scope        string `json:"scope"`
}

type ResourceOptionGoogleSheets struct {
	Authentication string                     `json:"authentication"`
	Options        *GoogleSheetsOAuth2Options `json:"opts"`
}

func NewResourceOptionGoogleSheetsByResource(resource *Resource) (*ResourceOptionGoogleSheets, error) {
	fmt.Printf("[DUMP] NewResourceOptionGoogleSheetsByResource().resource.Options: %+v\n", resource.Options)
	resourceOptionGoogleSheets := &ResourceOptionGoogleSheets{}
	resourceOptions := resource.ExportOptionsInMap()
	errInDecode := mapstructure.Decode(resourceOptions, &resourceOptionGoogleSheets)
	if errInDecode != nil {
		return nil, errInDecode
	}
	opts := &GoogleSheetsOAuth2Options{}
	errInDecodeSub := mapstructure.Decode(resourceOptions["opts"], &opts)
	if errInDecodeSub != nil {
		return nil, errInDecodeSub
	}
	resourceOptionGoogleSheets.Options = opts
	jstr, _ := json.Marshal(resourceOptionGoogleSheets)
	fmt.Printf("[DUMP] NewResourceOptionGoogleSheetsByResource().resourceOptionGoogleSheets: %+v\n", string(jstr))
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

func (i *ResourceOptionGoogleSheets) UpdateByExchangeTokenResponse(repo *oauthgoogle.ExchangeTokenResponse) {
	i.Options.AccessToken = repo.AccessToken
	i.Options.RefreshToken = repo.RefreshToken
	i.Options.ExpiresIn = repo.Expiry
	i.Options.Scope = repo.Scope
	i.Options.TokenType = repo.TokenType
}

func (i *ResourceOptionGoogleSheets) ExportInString() string {
	byteData, _ := json.Marshal(i)
	return string(byteData)
}
