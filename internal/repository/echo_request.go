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
	"fmt"
)

const (
	ECHO_REQ_DEFAULT_MODEL  = "gpt-3.5-turbo"
	ECHO_REQ_DEFAULT_TOKENS = 1000
)

const (
	ECHO_MSG_DEFAULT_ROLE = "user"
)

type EchoRequest struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	Messages  []*HistoryMessage `json:"messages"`
}

func (m *EchoRequest) SetMessages(msgs []*HistoryMessage) {
	m.Messages = append(m.Messages, msgs...)
}

func (m *EchoRequest) AppendMessage(msg *HistoryMessage) {
	m.Messages = append(m.Messages, msg)
}

func (m *EchoRequest) Export() string {
	r, _ := json.Marshal(m)
	fmt.Printf("[DUMP] EchoRequest: %+v\n", string(r))
	return string(r)
}

func NewEchoRequest() *EchoRequest {
	return &EchoRequest{
		Model:     ECHO_REQ_DEFAULT_MODEL,
		MaxTokens: ECHO_REQ_DEFAULT_TOKENS,
	}
}
