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

package repository

import (
	"encoding/json"
	"errors"
)

const APP_CONFIG_FIELD_PUBLIC = "public"

type AppConfig struct {
	Public bool `json:"public"` // switch for public app (which can view by anonymous user)
}

func (ac *AppConfig) ExportToJSONString() string {
	r, _ := json.Marshal(ac)
	return string(r)
}

func (ac *AppConfig) IsPublic() bool {
	return ac.Public
}

func NewAppConfigByConfigAppRawRequest(rawReq map[string]interface{}) (*AppConfig, error) {
	var assertPass bool
	appConfig := &AppConfig{}
	for key, value := range rawReq {
		switch key {
		case APP_CONFIG_FIELD_PUBLIC:
			appConfig.Public, assertPass = value.(bool)
			if !assertPass {
				return nil, errors.New("update app config failed due to assert failed.")
			}
		default:
		}
	}
	return appConfig, nil
}
