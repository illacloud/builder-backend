package oauthgoogle

import "encoding/json"

const (
	GOOGLE_OAUTH2_API = "https://oauth2.googleapis.com/token"
)

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	Expiry      int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func NewRefreshTokenResponse() *RefreshTokenResponse {
	return &RefreshTokenResponse{}
}

func (resp *RefreshTokenResponse) ExportAccessToken() string {
	return resp.AccessToken
}

func RefreshOAuthToken(refreshToken string) (*RefreshTokenSuccessResponse, error) {
	conf := config.GetInstance()
	googleOAuthClientID := conf.GetIllaGoogleSheetsClientID()
	googleOAuthClientSecret := conf.GetIllaGoogleSheetsClientSecret()
	client := resty.New()
	// request
	resp, errInPost := client.R().
		SetFormData(map[string]string{
			"client_id":     googleOAuthClientID,
			"client_secret": googleOAuthClientSecret,
			"refresh_token": refreshToken,
			"grant_type":    "refresh_token",
		}).
		Post(GOOGLE_OAUTH2_API)

	if resp.IsError() {
		return nil, errInPost
	}
	// unmarshal
	refreshTokenResponse := RefreshTokenResponse()
	errInUnmarshal := json.Unmarshal(resp.Body(), &refreshTokenResponse)
	if errInUnmarshal != nil {
		return nil, errInUnmarshal
	}
	return refreshTokenResponse, nil
}
