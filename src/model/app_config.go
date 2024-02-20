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
const APP_CONFIG_FIELD_PUBLISHED_TO_MARKETPLACE = "publishedToMarketplace"
const APP_CONFIG_FIELD_PUBLISH_WITH_AI_AGENT = "publishWithAIAgent"
const APP_CONFIG_FIELD_APP_TYPE = "appType"

const (
	APP_TYPE_PC     = 1
	APP_TYPE_MOBILE = 2
)

const (
	APP_TYPE_STRING_PC     = "pc"
	APP_TYPE_STRING_MOBILE = "mobile"
)

var appTypeMap = map[string]int{
	APP_TYPE_STRING_PC:     APP_TYPE_PC,
	APP_TYPE_STRING_MOBILE: APP_TYPE_MOBILE,
}

var appTypeToStringMap = map[int]string{
	APP_TYPE_PC:     APP_TYPE_STRING_PC,
	APP_TYPE_MOBILE: APP_TYPE_STRING_MOBILE,
}

type AppConfig struct {
	Public                 bool   `json:"public"` // switch for public app (which can view by anonymous user)
	WaterMark              bool   `json:"waterMark"`
	Description            string `json:"description"`
	PublishedToMarketplace bool   `json:"publishedToMarketplace"`
	PublishWithAIAgent     bool   `json:"publishWithAIAgent"`
	Cover                  string `json:"cover"`
	AppType                int    `json:"appType"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Public:      false,
		WaterMark:   true,
		Description: "",
		AppType:     APP_TYPE_PC,
	}
}

func (appConfig *AppConfig) ExportToJSONString() string {
	r, _ := json.Marshal(appConfig)
	return string(r)
}

func (appConfig *AppConfig) ExportAppType() int {
	if appConfig.AppType == 0 {
		return APP_TYPE_PC // default type is pc
	}
	return appConfig.AppType
}

func (appConfig *AppConfig) ExportAppTypeToString() string {
	if appConfig.AppType == 0 {
		return APP_TYPE_STRING_PC // default type is pc
	}
	return appTypeToStringMap[appConfig.AppType]
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
	appConfig.PublishedToMarketplace = true
}

func (appConfig *AppConfig) SetNotPublishedToMarketplace() {
	appConfig.PublishedToMarketplace = false
}

func (appConfig *AppConfig) SetPublishWithAIAgent() {
	appConfig.PublishWithAIAgent = true
}

func (appConfig *AppConfig) SetNotPublishWithAIAgent() {
	appConfig.PublishWithAIAgent = false
}

func (appConfig *AppConfig) SetCover(cover string) {
	appConfig.Cover = cover
}

func (appConfig *AppConfig) SetAppType(appType int) {
	appConfig.AppType = appType
}

func (appConfig *AppConfig) SetAppTypeByString(appType string) error {
	hit := false
	appConfig.AppType, hit = appTypeMap[appType]
	if !hit {
		return errors.New("invalied app type")
	}
	return nil
}

func (appConfig *AppConfig) UpdateAppConfigByConfigAppRawRequest(rawReq map[string]interface{}) error {
	assertPass := true
	for key, value := range rawReq {
		switch key {
		case APP_CONFIG_FIELD_PUBLIC:
			appConfig.Public, assertPass = value.(bool)
			if !assertPass {
				return errors.New("update app config failed due to assert failed")
			}
		case APP_CONFIG_FIELD_WATER_MARK:
			appConfig.WaterMark, assertPass = value.(bool)
			if !assertPass {
				return errors.New("update app config failed due to assert failed")
			}
		case APP_CONFIG_FIELD_DESCRIPTION:
			appConfig.Description, assertPass = value.(string)
			if !assertPass {
				return errors.New("update app config failed due to assert failed")
			}
		case APP_CONFIG_FIELD_PUBLISHED_TO_MARKETPLACE:
			appConfig.PublishedToMarketplace, assertPass = value.(bool)
			if !assertPass {
				return errors.New("update app config failed due to assert failed")
			}
		case APP_CONFIG_FIELD_PUBLISH_WITH_AI_AGENT:
			appConfig.PublishWithAIAgent, assertPass = value.(bool)
			if !assertPass {
				return errors.New("update app config failed due to assert failed")
			}
		case APP_CONFIG_FIELD_APP_TYPE:
			appTypeInString, assertPass := value.(string)
			if !assertPass {
				return errors.New("update app config failed due to assert failed")
			}
			appType, hit := appTypeMap[appTypeInString]
			if !hit {
				return errors.New("update app config failed due to invalied app type")
			}
			appConfig.AppType = appType
		default:
		}
	}
	// check app config phrase
	if appConfig.PublishedToMarketplace && !appConfig.Public {
		return errors.New("can not make app to private, this app already published to marketplace")
	}
	return nil
}

func NewAppConfigByDefault() *AppConfig {
	return &AppConfig{
		Public:      false,
		WaterMark:   true,
		Description: "",
		AppType:     APP_TYPE_PC,
	}
}
