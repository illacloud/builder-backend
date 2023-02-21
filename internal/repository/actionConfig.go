// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by actionlicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"encoding/json"
	"errors"
)

const ACTION_CONFIG_FIELD_PUBLIC = "public"

type ActionConfig struct {
	Public bool `json:"public"` // switch for public action (which can view by anonymous user)
}

func NewActionConfig() *ActionConfig {
	return &ActionConfig{
		Public: false,
	}
}

func (ac *ActionConfig) ExportToJSONString() string {
	r, _ := json.Marshal(ac)
	return string(r)
}

func (ac *ActionConfig) SetPublic() {
	ac.Public = true
}

func (ac *ActionConfig) SetPrivate() {
	ac.Public = false
}

func NewActionConfigByConfigAppRawRequest(rawReq map[string]interface{}) (*ActionConfig, error) {
	var assertPass bool
	actionConfig := &ActionConfig{}
	for key, value := range rawReq {
		switch key {
		case ACTION_CONFIG_FIELD_PUBLIC:
			actionConfig.Public, assertPass = value.(bool)
			if !assertPass {
				return nil, errors.New("update action config failed due to assert failed.")
			}
		default:
		}
	}
	return actionConfig, nil
}

func NewActionConfigByConfigActionRawRequest(rawReq map[string]interface{}) (*ActionConfig, error) {
	var assertPass bool
	actionConfig := &ActionConfig{}
	for key, value := range rawReq {
		switch key {
		case ACTION_CONFIG_FIELD_PUBLIC:
			actionConfig.Public, assertPass = value.(bool)
			if !assertPass {
				return nil, errors.New("update action config failed due to assert failed.")
			}
		default:
		}
	}
	return actionConfig, nil
}
