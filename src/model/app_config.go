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

package model

import (
	"encoding/json"
	"errors"
)

const APP_CONFIG_FIELD_PUBLIC = "public"
const APP_CONFIG_FIELD_WATER_MARK = "waterMark"
const APP_CONFIG_FIELD_DESCRIPTION = "description"

type AppConfig struct {
	Public                 bool   `json:"public"` // switch for public app (which can view by anonymous user)
	WaterMark              bool   `json:"waterMark"`
	Description            string `json:"description"`
	publishedToMarketplace bool   `json:"publishedToMarketplace"`
	cover                  string `json:"cover"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Public:      false,
		WaterMark:   true,
		Description: "",
	}
}

func (appConfig *AppConfig) ExportToJSONString() string {
	r, _ := json.Marshal(appConfig)
	return string(r)
}

func (appConfig *AppConfig) IsPublic() bool {
	return appConfig.Public
}

func (appConfig *AppConfig) EnableWaterMark() {
	appConfig.WaterMark = true
}

func (appConfig *AppConfig) DisableWaterMark() {
	appConfig.WaterMark = false
}

func (appConfig *AppConfig) SetublishedToMarketplace() {
	appConfig.publishedToMarketplace = true
}

func (appConfig *AppConfig) SetNotPublishedToMarketplace() {
	appConfig.publishedToMarketplace = false
}

func (appConfig *AppConfig) UpdateAppConfigByConfigAppRawRequest(rawReq map[string]interface{}) error {
	var assertPass bool
	for key, value := range rawReq {
		switch key {
		case APP_CONFIG_FIELD_PUBLIC:
			appConfig.Public, assertPass = value.(bool)
			if !assertPass {
				return errors.New("update app config failed due to assert failed.")
			}
		case APP_CONFIG_FIELD_WATER_MARK:
			appConfig.WaterMark, assertPass = value.(bool)
			if !assertPass {
				return errors.New("update app config failed due to assert failed.")
			}
		case APP_CONFIG_FIELD_DESCRIPTION:
			appConfig.Description, assertPass = value.(string)
			if !assertPass {
				return errors.New("update app config failed due to assert failed.")
			}
		default:
		}
	}
	return nil
}

func NewAppConfigByDefault() *AppConfig {
	return &AppConfig{
		Public:      false,
		WaterMark:   true,
		Description: "",
	}
}
