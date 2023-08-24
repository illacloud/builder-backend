package oauthgithub

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/illacloud/illa-resource-manager-backend/src/utils/config"
	"github.com/illacloud/illa-resource-manager-backend/src/utils/randomstring"
)

const GITHUB_OAUTH_AUTHORIZE_URI_TEMPLATE = "https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s"
const GITHUB_OAUTH_ACCESS_TOKEN_URI_TEMPLATE = "https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s"
const GITHUB_OAUTH_URI_TEMPLATE = "https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s"
const GITHUB_USER_INFO_API = "https://api.github.com/user"
const GITHUB_USER_EMAIL_API = "https://api.github.com/user/emails"
const GITHUB_DEFAULT_SCOPE = "read:user user:email"
const STATE_LENGTH = 16

type GithubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectUrl  string
	Scope        string
	State        string // random string for against cross-site request forgery attacks.
}

type GithubOAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func NewGithubOAuthToken() *GithubOAuthToken {
	return &GithubOAuthToken{}
}

func (t *GithubOAuthToken) ExportToken() string {
	return t.AccessToken
}

func NewGithubOAuthConfig(config *config.Config, redirectURL string, landing string) *GithubOAuthConfig {
	state, _ := json.Marshal(map[string]interface{}{"token": randomstring.RandStringBytesMaskImpr(STATE_LENGTH), "landing": landing})
	return &GithubOAuthConfig{
		ClientID:     config.GetGithubOAuthClientID(),
		ClientSecret: config.GetGithubOAuthClientSecret(),
		RedirectUrl:  redirectURL,
		Scope:        GITHUB_DEFAULT_SCOPE,
		State:        string(state),
	}
}

func GenerateOAuthAuthorizeURI(config *GithubOAuthConfig) string {
	return fmt.Sprintf(GITHUB_OAUTH_AUTHORIZE_URI_TEMPLATE, url.QueryEscape(config.ClientID), url.QueryEscape(config.RedirectUrl), url.QueryEscape(config.Scope), url.QueryEscape(config.State))
}

func GenerateOAuthAccessTokenURI(config *GithubOAuthConfig, code string) string {
	return fmt.Sprintf(GITHUB_OAUTH_ACCESS_TOKEN_URI_TEMPLATE, url.QueryEscape(config.ClientID), url.QueryEscape(config.ClientSecret), url.QueryEscape(code))
}

func Exchange(config *GithubOAuthConfig, code string) (*GithubOAuthToken, error) {
	// get exchange uri
	exchanegURI := GenerateOAuthAccessTokenURI(config, code)
	// get token
	var req *http.Request
	generalError := errors.New("github oauth exchange failed.") // @note: this general error avoid oauth client secret leak.
	var err error
	if req, err = http.NewRequest(http.MethodGet, exchanegURI, nil); err != nil {
		return nil, generalError
	}
	req.Header.Set("accept", "application/json")
	var httpClient = http.Client{}
	var res *http.Response
	if res, err = httpClient.Do(req); err != nil {
		return nil, generalError
	}

	// parse token
	token := NewGithubOAuthToken()
	if err = json.NewDecoder(res.Body).Decode(token); err != nil {
		return nil, generalError
	}
	return token, nil
}

func GetUserInfo(token *GithubOAuthToken) (map[string]interface{}, error) {
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, GITHUB_USER_INFO_API, nil); err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token.AccessToken))

	var client = http.Client{}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}

	// decode respoonse
	var userInfo = make(map[string]interface{})
	if err = json.NewDecoder(res.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return userInfo, nil
}

func ExtractUsername(userInfo map[string]interface{}) (string, error) {
	nameRaw, hit := userInfo["name"]
	if !hit {
		return "", errors.New("can not get field name from user info.")
	}
	name, errInCast := nameRaw.(string)
	if !errInCast {
		return "", errors.New("can not get field name from user info.")
	}
	return name, nil
}

func GetUserEmail(token *GithubOAuthToken) (string, error) {
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, GITHUB_USER_EMAIL_API, nil); err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	var client = http.Client{}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return "", err
	}

	// the paylaod be like:
	//
	//	userEmail: [
	//				map[email:example-a@gmail.com primary:false verified:true visibility:<nil>]
	//				map[email:example-b@outlook.com primary:true verified:true visibility:public]
	//				map[email:example-c@outlook.com primary:false verified:true visibility:<nil>]
	//				map[email:example-d@outlook.com primary:false verified:true visibility:<nil>]
	//			]
	var emails = make([]interface{}, 0)
	if err = json.NewDecoder(res.Body).Decode(&emails); err != nil {
		return "", err
	}

	// extract email
	if len(emails) == 0 {
		return "", errors.New("invalied oauth email info.")
	}
	for _, info := range emails {
		infoCasted, castInfoSuccess := info.(map[string]interface{})
		if !castInfoSuccess {
			return "", errors.New("invalied oauth user info in cast.")
		}
		// check verified
		if infoCasted["verified"] == false {
			continue
		}
		if infoCasted["primary"] == false {
			continue
		}
		email, castEmailSuccess := infoCasted["email"].(string)
		if !castEmailSuccess {
			return "", errors.New("invalied oauth user info in cast email.")
		}
		return email, nil
	}

	return "", errors.New("have no avaliable oauth email.")

}
