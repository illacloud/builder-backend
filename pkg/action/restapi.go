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

package action

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/illa-family/builder-backend/pkg/connector"

	"github.com/mitchellh/mapstructure"
)

const (
	METHOD_GET    = "GET"
	METHOD_POST   = "POST"
	METHOD_PUT    = "PUT"
	METHOD_DELETE = "DELETE"
	METHOD_PATCH  = "PATCH"

	BODY_JSON   = "json"
	BODY_RAW    = "raw"
	BODY_FORM   = "form-data"
	BODY_NONE   = "none"
	BODY_BINARY = "binary"
	BODY_XWFU   = "x-www-form-urlencoded"

	AUTH_BASIC  = "basic"
	AUTH_BEARER = "bearer"
)

type RestApiAction struct {
	Type            string
	RestApiTemplate RestApiTemplate
	Resource        *connector.Connector
}

type RestApiTemplate struct {
	Url       string
	Method    string
	UrlParams [][]string
	Headers   [][]string
	BodyType  string
	Body      [][]string
	Cookies   [][]string
}

type RestApiResource struct {
	Headers               [][]string
	Authentication        string
	AuthenticationOptions AuthenticationOptions
}

type AuthenticationOptions struct {
	BasicUsername string
	BasicPassword string
	BearerToken   string
}

func (r *RestApiAction) Run() (interface{}, error) {
	client := &http.Client{}

	reqUrl := r.RestApiTemplate.Url

	var reqBody io.Reader
	switch r.RestApiTemplate.BodyType {
	case BODY_JSON:
		bodyJson := map[string]interface{}{}
		for _, kv := range r.RestApiTemplate.Body {
			bodyJson[kv[0]] = kv[1]
		}
		bytesJson, _ := json.Marshal(bodyJson)
		reqBody = bytes.NewReader(bytesJson)
		break
	default:
		break
	}

	req, err := http.NewRequest(r.RestApiTemplate.Method, reqUrl, reqBody)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for _, param := range r.RestApiTemplate.UrlParams {
		if len(param) == 2 {
			query.Add(param[0], param[1])
		}
	}
	req.URL.RawQuery = query.Encode()

	if r.Resource != nil {
		var resourceApi RestApiResource
		mapstructure.Decode(r.Resource.Options, resourceApi)
		for _, header := range resourceApi.Headers {
			if len(header) == 2 {
				req.Header.Add(header[0], header[1])
			}
		}
		switch resourceApi.Authentication {
		case AUTH_BASIC:
			req.SetBasicAuth(resourceApi.AuthenticationOptions.BasicUsername, resourceApi.AuthenticationOptions.BasicPassword)
			break
		case AUTH_BEARER:
			bearer := "Bearer " + resourceApi.AuthenticationOptions.BearerToken
			req.Header.Add("Authorization", bearer)
			break
		default:
			break
		}
	}
	for _, header := range r.RestApiTemplate.Headers {
		if len(header) == 2 {
			req.Header.Add(header[0], header[1])
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("invalid response")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, err
	}

	return res, nil
}
