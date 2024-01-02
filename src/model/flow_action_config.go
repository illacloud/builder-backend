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

package model

import (
	"encoding/json"
)

type FlowActionConfig struct {
	IsVirtualResource  bool                `json:"isVirtualResource"`
	FlowAdvancedConfig *FlowAdvancedConfig `json:"advancedConfig"` // 2023_4_20: add advanced config for action
	FlowMockConfig     *FlowMockConfig     `json:"mockConfig"`
}

type FlowAdvancedConfig struct {
	Runtime            string   `json:"runtime"`
	Pages              []string `json:"pages"`
	DelayWhenLoaded    string   `json:"delayWhenLoaded"`
	DisplayLoadingPage bool     `json:"displayLoadingPage"`
	IsPeriodically     bool     `json:"isPeriodically"`
	PeriodInterval     string   `json:"periodInterval"`
	Mock               string   `json:"mock"`
}

func NewFlowActionConfig() *FlowActionConfig {
	return &FlowActionConfig{
		FlowAdvancedConfig: &FlowAdvancedConfig{
			Runtime:            "none",
			Pages:              []string{},
			DelayWhenLoaded:    "",
			DisplayLoadingPage: false,
			IsPeriodically:     false,
			PeriodInterval:     "",
		},
		FlowMockConfig: &FlowMockConfig{
			Enabled:  false,
			MockData: "",
		},
	}
}

func (ac *FlowActionConfig) ExportToJSONString() string {
	r, _ := json.Marshal(ac)
	return string(r)
}

func (ac *FlowActionConfig) SetIsVirtualResource() {
	ac.IsVirtualResource = true
}

func (ac *FlowActionConfig) SetIsNotVirtualResource() {
	ac.IsVirtualResource = false
}
