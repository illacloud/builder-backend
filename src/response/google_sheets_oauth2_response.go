package response

import (
	"net/url"

	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/config"
)

type GoogleSheetsOAuth2Response struct {
	URL string `json:"url"`
}

func NewGoogleSheetsOAuth2Response(accessType int, accessToken string) *GoogleSheetsOAuth2Response {
	conf := config.GetInstance()
	googleOAuthClientID := conf.GetIllaGoogleSheetsClientID()
	redirectURI := conf.GetIllaGoogleSheetsRedirectURI()
	urlObject := url.URL{}
	if accessType == model.GOOGLE_SHEETS_OAUTH2_ACCESS_TYPE_READ_AND_WRITE {
		urlObject = url.URL{
			Scheme:   "https",
			Host:     "accounts.google.com",
			Path:     "o/oauth2/v2/auth/oauthchooseaccount",
			RawQuery: "response_type=" + "code" + "&client_id=" + googleOAuthClientID + "&redirect_uri=" + redirectURI + "&state=" + accessToken + "&scope=" + "https://www.googleapis.com/auth/drive https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/spreadsheets" + "&access_type=" + "offline" + "&prompt=" + "consent" + "&service=" + "lso" + "&o2v=" + "2" + "&flowName=" + "GeneralOAuthFlow",
		}
	} else {
		urlObject = url.URL{
			Scheme:   "https",
			Host:     "accounts.google.com",
			Path:     "o/oauth2/v2/auth/oauthchooseaccount",
			RawQuery: "response_type=" + "code" + "&client_id=" + googleOAuthClientID + "&redirect_uri=" + redirectURI + "&state=" + accessToken + "&scope=" + "https://www.googleapis.com/auth/spreadsheets.readonly https://www.googleapis.com/auth/drive.readonly" + "&access_type=" + "offline" + "&prompt=" + "consent" + "&service=" + "lso" + "&o2v=" + "2" + "&flowName=" + "GeneralOAuthFlow",
		}
	}
	return &GoogleSheetsOAuth2Response{
		URL: urlObject.String(),
	}
}

func (resp *GoogleSheetsOAuth2Response) ExportForFeedback() interface{} {
	return resp
}
