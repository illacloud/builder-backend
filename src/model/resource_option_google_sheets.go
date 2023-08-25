package model

import "github.com/mitchellh/mapstructure"

const (
	GOOGLE_SHEET_OAUTH_TYPE = "oauth2"
)

type ResourceOptionGoogleSheets struct {
	Authentication string
	Opts           OAuth2Opts
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
