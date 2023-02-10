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

package graphql

import (
	"encoding/json"
	"errors"

	"github.com/illacloud/builder-backend/pkg/plugins/common"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	ResourceOpts Resource
	ActionOpts   Action
}

func (g *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &g.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate graphql options
	validate := validator.New()
	if err := validate.Struct(g.ResourceOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (g *Connector) ValidateActionOptions(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &g.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate graphql options
	validate := validator.New()
	if err := validate.Struct(g.ActionOpts); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	return common.ValidateResult{Valid: true}, nil
}

func (g *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &g.ResourceOpts); err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	queryParams := make(map[string]string)
	headers := make(map[string]string)
	cookies := make(map[string]string)
	for _, param := range g.ResourceOpts.URLParams {
		if param["key"] != "" {
			queryParams[param["key"]] = param["value"]
		}
	}

	for _, header := range g.ResourceOpts.Headers {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}

	for _, cookie := range g.ResourceOpts.Cookies {
		if cookie["key"] != "" {
			cookies[cookie["key"]] = cookie["value"]
		}
	}

	resp, err := g.doQuery(g.ResourceOpts.BaseURL, queryParams, headers, cookies, g.ResourceOpts.Authentication,
		g.ResourceOpts.AuthContent, "{__typename}", nil)
	if err != nil {
		return common.ConnectionResult{Success: false}, err
	}

	if resp.IsError() {
		return common.ConnectionResult{Success: false}, errors.New("unknown error")
	}

	return common.ConnectionResult{Success: true}, nil
}

func (g *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{
		Success: true,
		Schema:  nil,
	}, nil
}

func (g *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &g.ResourceOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format action options
	if err := mapstructure.Decode(actionOptions, &g.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	queryParams := make(map[string]string)
	headers := make(map[string]string)
	cookies := make(map[string]string)
	for _, param := range g.ResourceOpts.URLParams {
		if param["key"] != "" {
			queryParams[param["key"]] = param["value"]
		}
	}

	for _, header := range g.ResourceOpts.Headers {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}
	for _, header := range g.ActionOpts.Headers {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}

	for _, cookie := range g.ResourceOpts.Cookies {
		if cookie["key"] != "" {
			cookies[cookie["key"]] = cookie["value"]
		}
	}

	vars := make(map[string]interface{})
	for _, variable := range g.ActionOpts.Variables {
		if variable["key"] != "" {
			vars[variable["key"].(string)] = variable["value"]
		}
	}

	resp, err := g.doQuery(g.ResourceOpts.BaseURL, queryParams, headers, cookies, g.ResourceOpts.Authentication,
		g.ResourceOpts.AuthContent, g.ActionOpts.Query, vars)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	if resp.IsError() {
		return common.RuntimeResult{Success: false}, errors.New("unknown error")
	}
	body := make(map[string]interface{})
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	return common.RuntimeResult{Success: true, Rows: []map[string]interface{}{body}}, nil
}
