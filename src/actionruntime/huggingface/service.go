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
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/illacloud/builder-backend/src/actionruntime/common"

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

	return common.ValidateResult{Valid: true}, nil
}

func (h *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
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

func (h *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &h.ResourceOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format action options
	if err := mapstructure.Decode(actionOptions, &h.ActionOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// Create a Resty Client
	client := resty.New().R()
	// set Hugging Face token
	client.SetAuthToken(h.ResourceOpts.Token)
	// build Hugging Face request
	switch h.ActionOpts.Params.Inputs.Type {
	case INPUT_PAIRS_MODE:
		pairs := make([]Pairs, 0)
		if err := mapstructure.Decode(h.ActionOpts.Params.Inputs.Content, &pairs); err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		inputs := make(map[string]interface{})
		for _, pair := range pairs {
			if pair.Key != "" {
				inputs[pair.Key] = pair.Value
			}
		}
		params := make(map[string]interface{})
		if h.ActionOpts.Params.WithDetailParams {
			params = buildDetailedParams(h.ActionOpts.Params.DetailParams)
		}
		reqBody := make(map[string]interface{}, 2)
		reqBody["inputs"] = inputs
		if len(params) > 0 {
			reqBody["parameters"] = params
		}
		client.SetBody(reqBody)
		client.SetHeader("Content-Type", "application/json")
		break
	case INPUT_TEXT_MODE, INPUT_JSON_MODE:
		reqBody := make(map[string]interface{})
		reqBody["inputs"] = h.ActionOpts.Params.Inputs.Content
		params := make(map[string]interface{})
		if h.ActionOpts.Params.WithDetailParams {
			params = buildDetailedParams(h.ActionOpts.Params.DetailParams)
		}
		if len(params) > 0 {
			reqBody["parameters"] = params
		}
		client.SetBody(reqBody)
		client.SetHeader("Content-Type", "application/json")
		break
	case INPUT_BINARY_MODE:
		bs, _ := h.ActionOpts.Params.Inputs.Content.(string)
		binaryBytes, err := base64.StdEncoding.DecodeString(bs)
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		if err != nil {
			return common.RuntimeResult{Success: false}, err
		}
		client.SetBody(binaryBytes)
		break
	default:
		return common.RuntimeResult{}, errors.New("unsupported input parameters")
	}

	resp, err := client.Post(HF_API_ADDRESS + h.ActionOpts.ModelID)
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	body := make(map[string]interface{})
	listBody := make([]map[string]interface{}, 0)
	matrixBody := make([][]map[string]interface{}, 0)
	if err := json.Unmarshal(resp.Body(), &body); err == nil {
		res.Rows = append(res.Rows, body)
	}
	if err := json.Unmarshal(resp.Body(), &listBody); err == nil {
		res.Rows = listBody
	}
	if err := json.Unmarshal(resp.Body(), &matrixBody); err == nil && len(matrixBody) == 1 {
		res.Rows = matrixBody[0]
	}
	if !isBase64Encoded(string(resp.Body())) {
		res.Extra["raw"] = base64Encode(resp.Body())
	} else {
		res.Extra["raw"] = string(resp.Body())
	}
	res.Extra["headers"] = resp.Header()
	res.Extra["statusCode"] = resp.StatusCode()
	res.Extra["statusText"] = resp.Status()
	if err != nil {
		res.Success = false
		return res, err
	}
	res.Success = true

	return res, nil
}

func isBase64Encoded(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func base64Encode(s []byte) string {
	encoded := base64.StdEncoding.EncodeToString(s)
	return encoded
}
