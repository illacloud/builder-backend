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
	"fmt"
	"strings"

	"github.com/illacloud/builder-backend/src/actionruntime/common"
	parser_template "github.com/illacloud/builder-backend/src/utils/parser/template"

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

func (g *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
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

func (g *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &g.ResourceOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format action options
	if err := mapstructure.Decode(actionOptions, &g.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	getContext := func() map[string]interface{} {
		contextRaw := rawActionOptions["context"]
		context, _ := contextRaw.(map[string]interface{})
		return context
	}
	preprocessKVPairSliceWithContext := func(KVPairSlice []map[string]string, context map[string]interface{}) ([]map[string]string, error) {
		var errInPreprocessTemplate error
		if len(KVPairSlice) > 0 {
			newKVPairSlice := make([]map[string]string, 0)
			for _, kvpair := range KVPairSlice {
				newKVPair := map[string]string{
					"key":   "",
					"value": "",
				}
				// process key
				key, hitKye := kvpair["key"]
				if !hitKye {
					return nil, errors.New("preprocessKVPairSlice() can not find field \"key\"")
				}
				if newKVPair["key"], errInPreprocessTemplate = parser_template.AssembleTemplateWithVariable(key, context); errInPreprocessTemplate != nil {
					return nil, errInPreprocessTemplate
				}
				// process value
				value, hitValue := kvpair["value"]
				if !hitValue {
					return nil, errors.New("preprocessKVPairSlice() can not find value field \"value\"")
				}

				if newKVPair["value"], errInPreprocessTemplate = parser_template.AssembleTemplateWithVariable(value, context); errInPreprocessTemplate != nil {
					return nil, errInPreprocessTemplate
				}
				newKVPairSlice = append(newKVPairSlice, newKVPair)
			}
			return newKVPairSlice, nil
		}
		return nil, nil
	}
	preprocessKVPairSliceForGraphqlVariableWithContext := func(KVPairSlice []map[string]interface{}, context map[string]interface{}) ([]map[string]interface{}, error) {
		var errInPreprocessTemplate error
		if len(KVPairSlice) > 0 {
			newKVPairSlice := make([]map[string]interface{}, 0)
			for _, kvpair := range KVPairSlice {
				newKVPair := map[string]interface{}{
					"key":   "",
					"value": "",
				}
				// process key
				key, hitKye := kvpair["key"].(string)
				if !hitKye {
					return nil, errors.New("preprocessKVPairSlice() can not find field or target field is not string, field name: \"key\"")
				}
				if newKVPair["key"], errInPreprocessTemplate = parser_template.AssembleTemplateWithVariable(key, context); errInPreprocessTemplate != nil {
					return nil, errInPreprocessTemplate
				}
				// process value
				value, hitValue := kvpair["value"]
				if !hitValue {
					return nil, errors.New("preprocessKVPairSlice() can not find value field \"value\"")
				}
				valueAsserted, assertValuePass := value.(string)
				if !assertValuePass {
					// not a string value (or not string template), we use original value
					newKVPair["value"] = value
				} else {
					if newKVPair["value"], errInPreprocessTemplate = parser_template.AssembleTemplateWithVariable(valueAsserted, context); errInPreprocessTemplate != nil {
						return nil, errInPreprocessTemplate
					}

					// check if value itself are context key, then use context original value as value
					valueAssertedTrimmed := strings.TrimSpace(valueAsserted)
					// start with "{{" and end with "}}"
					if strings.Index(valueAssertedTrimmed, "{{") == 0 && strings.Index(valueAssertedTrimmed, "}}") == len(valueAssertedTrimmed)-2 {
						valueAssertedWithoutLeftBrace := strings.Replace(valueAsserted, "{{", "", -1)
						valueAssertedRawValueInBrace := strings.Replace(valueAssertedWithoutLeftBrace, "}}", "", -1)
						hitContext := false
						newKVPair["value"], hitContext = context[valueAssertedRawValueInBrace]
						if !hitContext {
							newKVPair["value"] = nil
						}
					}
				}
				newKVPairSlice = append(newKVPairSlice, newKVPair)
			}
			return newKVPairSlice, nil
		}
		return nil, nil
	}

	queryParams := make(map[string]string)
	headers := make(map[string]string)
	cookies := make(map[string]string)
	vars := make(map[string]interface{})

	// get context
	context := getContext()

	// preprocess URL params
	urlParamsPreprocessed, errInPreprocessURLParamKVPair := preprocessKVPairSliceWithContext(g.ResourceOpts.URLParams, context)
	if errInPreprocessURLParamKVPair != nil {
		return common.RuntimeResult{Success: false}, errInPreprocessURLParamKVPair
	}
	for _, param := range urlParamsPreprocessed {
		if param["key"] != "" {
			queryParams[param["key"]] = param["value"]
		}
	}
	fmt.Printf("[DUMP] queryParams: %+v\n", queryParams)

	// preprocess Header
	headersPreprocessed, errInPreprocessHeadersKVPair := preprocessKVPairSliceWithContext(g.ResourceOpts.Headers, context)
	if errInPreprocessHeadersKVPair != nil {
		return common.RuntimeResult{Success: false}, errInPreprocessHeadersKVPair
	}
	for _, header := range headersPreprocessed {
		if header["key"] != "" {
			headers[header["key"]] = header["value"]
		}
	}
	fmt.Printf("[DUMP] headers: %+v\n", headers)

	// preprocess cookie
	cookiesPreprocessed, errInPreprocessCookiesKVPair := preprocessKVPairSliceWithContext(g.ResourceOpts.Cookies, context)
	if errInPreprocessCookiesKVPair != nil {
		return common.RuntimeResult{Success: false}, errInPreprocessCookiesKVPair
	}
	for _, cookie := range cookiesPreprocessed {
		if cookie["key"] != "" {
			cookies[cookie["key"]] = cookie["value"]
		}
	}
	fmt.Printf("[DUMP] cookies: %+v\n", cookies)

	// preprocess variables, the variables should returns with original data type from JSON. When context concat with string, it will return in string type.
	variablesPreprocessed, errInPreprocessVariablesKVPair := preprocessKVPairSliceForGraphqlVariableWithContext(g.ActionOpts.Variables, context)
	if errInPreprocessVariablesKVPair != nil {
		return common.RuntimeResult{Success: false}, errInPreprocessVariablesKVPair
	}
	for _, variable := range variablesPreprocessed {
		if variable["key"] != "" {
			vars[variable["key"].(string)] = variable["value"]
		}
	}
	fmt.Printf("[DUMP] vars: %+v\n", vars)

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
