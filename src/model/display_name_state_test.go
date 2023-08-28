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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveDisplayNameByPayload(t *testing.T) {
	var displayName interface{} = "testDisplayName-01"
	displayNameStr := "testDisplayName-01"
	resolvedResult, err := ResolveDisplayNameByPayload(displayName)

	assert.Nil(t, err)
	assert.Equal(t, displayNameStr, resolvedResult, "ResolveDisplayNameByPayload() resolve failed")
}

func TestResolveDisplayNameStateByPayload(t *testing.T) {
	var displayNameSet interface{} = []interface{}{"testDisplayName-01", "testDisplayName-02", "testDisplayName-03"}
	displayNameSetSlice := DisplayNameState{"testDisplayName-01", "testDisplayName-02", "testDisplayName-03"}
	resolvedResult, err := ResolveDisplayNameStateByPayload(displayNameSet)

	assert.Nil(t, err)
	assert.Equal(t, displayNameSetSlice, resolvedResult, "ResolveDisplayNameStateByPayload() resolve failed")
}

func TestConstructDisplayNameStateForUpdateByPayload(t *testing.T) {
	serilizationData := `{    "before":"image1",    "after":"image_1"}`
	var dnsfu map[string]interface{}
	var dnsfuExcepted DisplayNameStateForUpdate
	dnsfuExcepted.Before = "image1"
	dnsfuExcepted.After = "image_1"
	err := json.Unmarshal([]byte(serilizationData), &dnsfu)
	assert.Nil(t, err)
	resolvedResult, err := ConstructDisplayNameStateForUpdateByPayload(dnsfu)
	assert.Nil(t, err)
	assert.Equal(t, &dnsfuExcepted, resolvedResult, "ConstructDisplayNameStateForUpdateByPayload() resolve failed")

}
