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
	"github.com/illacloud/builder-backend/pkg/plugins/common"
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

func (r *RESTAPIConnector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{Success: false}, errors.New("unsupported type: REST API")
}

func (r *RESTAPIConnector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	var err error

	actionURLParams := map[string]string{}
	for _, param := range r.Action.UrlParams {
		if param["key"] != "" {
			actionURLParams[param["key"]] = param["value"]
		}
	}

	headers := map[string]string{}
	for _, header := range r.Resource.Headers {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}
	for _, header := range r.Action.Headers {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}

	cookies := map[string]string{}
	for _, cookie := range r.Resource.Cookies {
		if cookie["key"] != "" {
			cookies[cookie["key"]] = cookie["value"]
		}
	}
	for _, cookie := range r.Action.Cookies {
		if cookie["key"] != "" {
			cookies[cookie["key"]] = cookie["value"]
		}
	}

	client := resty.New()

	// get baseurl
	uri, err := url.Parse(r.Resource.BaseURL)
	if err != nil {
		res.Success = false
		return res, err
	}
	params := url.Values{}
	for _, v := range r.Resource.URLParams {
		if v["key"] != "" {
			params.Set(v["key"], v["value"])
		}
	}
	uri.RawQuery = params.Encode()
	baseURL := uri.String()

	// resty client set `resource` options
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
	actionClient.SetHeaders(headers)
	// set cookies, will override `resource` cookies
	actionCookies := make([]*http.Cookie, 0, len(r.Action.Cookies))
	for k, v := range cookies {
		actionCookies = append(actionCookies, &http.Cookie{Name: k, Value: v})
	}
	actionClient.SetCookies(actionCookies)

	// set body for action client
	switch r.Action.BodyType {
	case BODY_RAW:
		b := r.Action.ReflectBodyToRaw()
		actionClient.SetBody(b.Content)
		break
	case BODY_BINARY:
		b := r.Action.ReflectBodyToBinary()
		actionClient.SetBody(b)
		break
	case BODY_FORM:
		b := r.Action.ReflectBodyToMap()
		actionClient.SetBody(b)
		break
	case BODY_XWFU:
		b := r.Action.ReflectBodyToMap()
		actionClient.SetFormData(b)
		break
	case BODY_NONE:
		break
	}

	switch r.Action.Method {
	case METHOD_GET:
		resp, err := actionClient.SetQueryParams(actionURLParams).Get(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		listBody := make([]map[string]interface{}, 0)
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		if err := json.Unmarshal(resp.Body(), &listBody); err == nil {
			res.Rows = listBody
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_POST:
		resp, err := actionClient.SetQueryParams(actionURLParams).Post(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		listBody := make([]map[string]interface{}, 0)
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		if err := json.Unmarshal(resp.Body(), &listBody); err == nil {
			res.Rows = listBody
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_PUT:
		resp, err := actionClient.SetQueryParams(actionURLParams).Put(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		listBody := make([]map[string]interface{}, 0)
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		if err := json.Unmarshal(resp.Body(), &listBody); err == nil {
			res.Rows = listBody
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_PATCH:
		resp, err := actionClient.SetQueryParams(actionURLParams).Patch(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		listBody := make([]map[string]interface{}, 0)
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		if err := json.Unmarshal(resp.Body(), &listBody); err == nil {
			res.Rows = listBody
		}
		res.Extra["raw"] = resp.Body()
		res.Extra["headers"] = resp.Header()
		if err != nil {
			res.Success = false
			return res, err
		}
		break
	case METHOD_DELETE:
		resp, err := actionClient.SetQueryParams(actionURLParams).Delete(baseURL + r.Action.URL)
		body := make(map[string]interface{})
		listBody := make([]map[string]interface{}, 0)
		if err := json.Unmarshal(resp.Body(), &body); err == nil {
			res.Rows = append(res.Rows, body)
		}
		if err := json.Unmarshal(resp.Body(), &listBody); err == nil {
			res.Rows = listBody
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
