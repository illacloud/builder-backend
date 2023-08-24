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
	"errors"
)

type DisplayNameState []string

type DisplayNameStateForUpdate struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

func ResolveDisplayNameByPayload(data interface{}) (string, error) {
	var udata string
	var ok bool

	if udata, ok = data.(string); !ok {
		return "", errors.New("ResolveDisplayNameByPayload() failed, please check payload syntax.")
	}
	return udata, nil
}

func ResolveDisplayNameStateByPayload(data interface{}) (DisplayNameState, error) {
	var udata []interface{}
	var ok bool
	var dns DisplayNameState

	if udata, ok = data.([]interface{}); !ok {
		return nil, errors.New("ConstructDisplayNameByMap() failed, please check payload syntax.")
	}
	for _, v := range udata {
		dns = append(dns, v.(string))
	}
	return dns, nil
}

func ConstructDisplayNameStateForUpdateByPayload(data interface{}) (*DisplayNameStateForUpdate, error) {
	var udata map[string]interface{}
	var ok bool
	var dnsfu DisplayNameStateForUpdate

	if udata, ok = data.(map[string]interface{}); !ok {
		return nil, errors.New("ConstructDisplayNameStateForUpdateByPayload() failed, please check payload syntax.")
	}

	for k, v := range udata {
		switch k {
		case "before":
			dnsfu.Before = v.(string)
		case "after":
			dnsfu.After = v.(string)
		}
	}
	return &dnsfu, nil
}
