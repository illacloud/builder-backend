// Copyright 2023 Illa Soft, Inc.
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

package huggingface

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/illacloud/builder-backend/pkg/plugins/common"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	ResourceOpts Resource
	ActionOpts   Action
}

func (h *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &h.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Hugging Face options
	validate := validator.New()
	if err := validate.Struct(h.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Hugging Face token
	switch h.ResourceOpts.Authentication {
	case AUTH_NONE:
		return common.ValidateResult{Valid: false}, errors.New("authentication error")
	case AUTH_BASIC:
		return common.ValidateResult{Valid: false}, errors.New("unsupported authentication")
	case AUTH_BEARER:
		bearerToken, ok := h.ResourceOpts.AuthContent["token"]
		if !ok || bearerToken == "" {
			return common.ValidateResult{Valid: false}, errors.New("missing Hugging Face token")
		}
	default:
		return common.ValidateResult{Valid: false}, errors.New("authentication error")
	}
	return common.ValidateResult{Valid: true}, nil
}

func (h *Connector) ValidateActionOptions(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &h.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Hugging Face options
	validate := validator.New()
	if err := validate.Struct(h.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (h *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	return common.ConnectionResult{Success: false}, errors.New("unsupported type: Hugging Face")
}

func (h *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{Success: false}, errors.New("unsupported type: Hugging Face")
}

func (h *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	var err error

	actionURLParams := map[string]string{}
	for _, param := range h.ActionOpts.URLParams {
		if param["key"] != "" {
			actionURLParams[param["key"]] = param["value"]
		}
	}

	headers := map[string]string{}
	for _, header := range h.ResourceOpts.Headers {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}
	for _, header := range h.ActionOpts.Headers {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}

	cookies := map[string]string{}
	for _, cookie := range h.ResourceOpts.Cookies {
		if cookie["key"] != "" {
			cookies[cookie["key"]] = cookie["value"]
		}
	}
	for _, cookie := range h.ActionOpts.Cookies {
		if cookie["key"] != "" {
			cookies[cookie["key"]] = cookie["value"]
		}
	}

	client := resty.New()

	// get baseurl
	uri, err := url.Parse(h.ResourceOpts.BaseURL)
	if err != nil {
		res.Success = false
		return res, err
	}
	params := url.Values{}
	for _, v := range h.ResourceOpts.URLParams {
		if v["key"] != "" {
			params.Set(v["key"], v["value"])
		}
	}
	uri.RawQuery = params.Encode()
	baseURL := uri.String()

	// resty client set `resource` options
	// set auth
	switch h.ResourceOpts.Authentication {
	case AUTH_BASIC:
		break
	case AUTH_BEARER:
		client.SetAuthToken(h.ResourceOpts.AuthContent["token"])
	default:
		break
	}

	// resty client instance set `action` options
	actionClient := client.R()
	// set headers, will override `resource` headers
	actionClient.SetHeaders(headers)
	// set cookies, will override `resource` cookies
	actionCookies := make([]*http.Cookie, 0, len(h.ActionOpts.Cookies))
	for k, v := range cookies {
		actionCookies = append(actionCookies, &http.Cookie{Name: k, Value: v})
	}
	actionClient.SetCookies(actionCookies)

	// set body for action client
	switch h.ActionOpts.BodyType {
	case BODY_RAW:
		b := h.ActionOpts.ReflectBodyToRaw()
		actionClient.SetBody(b.Content)
		break
	case BODY_BINARY:
		b := h.ActionOpts.ReflectBodyToBinary()
		actionClient.SetBody(b)
		break
	case BODY_FORM:
		b := h.ActionOpts.ReflectBodyToMap()
		actionClient.SetBody(b)
		break
	case BODY_XWFU:
		b := h.ActionOpts.ReflectBodyToMap()
		actionClient.SetFormData(b)
		break
	case BODY_NONE:
		break
	}

	switch h.ActionOpts.Method {
	case METHOD_GET:
		break
	case METHOD_POST:
		resp, err := actionClient.SetQueryParams(actionURLParams).Post(baseURL + h.ActionOpts.URL)
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
		break
	case METHOD_PATCH:
		break
	case METHOD_DELETE:
		break
	}

	res.Success = true
	return res, nil
}
