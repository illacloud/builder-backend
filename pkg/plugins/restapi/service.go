// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package restapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/illa-family/builder-backend/pkg/plugins/common"
	"github.com/mitchellh/mapstructure"
)

type RESTAPIConnector struct {
	Resource RESTOptions
	Action   RESTTemplate
}

func (r *RESTAPIConnector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &r.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate restapi options
	validate := validator.New()
	if err := validate.Struct(r.Resource); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate restapi auth options
	switch r.Resource.Authentication {
	case AUTH_NONE:
		break
	case AUTH_BASIC:
		basicUsername, ok := r.Resource.AuthContent["username"]
		if !ok || basicUsername == "" {
			return common.ValidateResult{Valid: false}, errors.New("missing basic username")
		}
		basicPassword, ok := r.Resource.AuthContent["password"]
		if !ok || basicPassword == "" {
			return common.ValidateResult{Valid: false}, errors.New("missing basic password")
		}
		break
	case AUTH_BEARER:
		bearerToken, ok := r.Resource.AuthContent["token"]
		if !ok || bearerToken == "" {
			return common.ValidateResult{Valid: false}, errors.New("missing bearer token")
		}
		break
	}
	return common.ValidateResult{Valid: true}, nil
}

func (r *RESTAPIConnector) ValidateActionOptions(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format sql options
	if err := mapstructure.Decode(actionOptions, &r.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate restapi options
	validate := validator.New()
	if err := validate.Struct(r.Action); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (r *RESTAPIConnector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	return common.ConnectionResult{Success: false}, errors.New("unsupported type: REST API")
}

func (r *RESTAPIConnector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{}
	var err error

	client := resty.New()

	// get baseurl
	uri, err := url.Parse(r.Resource.BaseURL)
	if err != nil {
		res.Success = false
		return res, err
	}
	params := url.Values{}
	for k, v := range r.Resource.URLParams {
		params.Set(k, v)
	}
	uri.RawQuery = params.Encode()
	baseURL := uri.String()

	// resty client set `resource` options
	// set headers, can be overridden by action options
	client.SetHeaders(r.Resource.Headers)
	// set cookies, can be overridden by action options
	cookies := make([]*http.Cookie, 0, len(r.Resource.Cookies))
	for k, v := range r.Resource.Cookies {
		cookies = append(cookies, &http.Cookie{Name: k, Value: v})
	}
	client.SetCookies(cookies)
	// set auth
	switch r.Resource.Authentication {
	case AUTH_BASIC:
		client.SetBasicAuth(r.Resource.AuthContent["username"], r.Resource.AuthContent["password"])
		break
	case AUTH_BEARER:
		client.SetAuthToken(r.Resource.AuthContent["token"])
		break
	}

	// resty client instance set `action` options
	actionClient := client.R()
	// set headers, will override `resource` headers
	actionClient.SetHeaders(r.Action.Headers)
	// set cookies, will override `resource` cookies
	actionCookies := make([]*http.Cookie, 0, len(r.Action.Cookies))
	for k, v := range r.Action.Cookies {
		cookies = append(cookies, &http.Cookie{Name: k, Value: v})
	}
	actionClient.SetCookies(actionCookies)

	// set body for action client
	switch r.Action.BodyType {
	case BODY_JSON:
		actionClient.SetHeader("Content-Type", "application/json")
		actionClient.SetBody(r.Action.Body)
		break
	case BODY_FORM:
		actionClient.SetHeader("Content-Type", "application/form-data")
		actionClient.SetBody(r.Action.Body)
		break
	case BODY_XWFU:
		actionClient.SetFormData(r.Action.Body)
		break
	case BODY_NONE:
		break
	}

	switch r.Action.Method {
	case METHOD_GET:
		resp, err := actionClient.SetQueryParams(r.Action.UrlParams).Get(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_POST:
		resp, err := actionClient.SetQueryParams(r.Action.UrlParams).Post(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_PUT:
		resp, err := actionClient.SetQueryParams(r.Action.UrlParams).Put(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_PATCH:
		resp, err := actionClient.SetQueryParams(r.Action.UrlParams).Patch(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_DELETE:
		resp, err := actionClient.SetQueryParams(r.Action.UrlParams).Delete(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	}

	res.Success = true
	return res, nil
}
