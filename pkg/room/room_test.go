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

package room

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSample(t *testing.T) {
	assert.Nil(t, nil)
}

func TestGetServerAddress(t *testing.T) {
	addr0 := getServerAddress()
	assert.Equal(t, "localhost", addr0, "the server address should be localhost")

	os.Setenv("SERVER_ADDRESS", "127.0.0.1")
	addr1 := getServerAddress()
	assert.Equal(t, "127.0.0.1", addr1, "the server address should be 127.0.0.1")

	os.Setenv("SERVER_ADDRESS", "thisismyhost.com")
	addr2 := getServerAddress()
	assert.Equal(t, "thisismyhost.com", addr2, "the server address should be thisismyhost.com")

}
