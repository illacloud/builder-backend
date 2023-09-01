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

package websocket

import "encoding/json"

// for feedback
const ERROR_CODE_OK = 0
const ERROR_CODE_FAILED = 1
const ERROR_CODE_NEED_ENTER = 2
const ERROR_CODE_BROADCAST = 0
const ERROR_CODE_PONG = 3
const ERROR_CODE_LOGGEDIN = 0
const ERROR_CODE_LOGIN_FAILED = 4
const ERROR_CREATE_STATE_OK = 0
const ERROR_CREATE_STATE_FAILED = 5
const ERROR_DELETE_STATE_OK = 0
const ERROR_DELETE_STATE_FAILED = 6
const ERROR_UPDATE_STATE_OK = 0
const ERROR_UPDATE_STATE_FAILED = 7
const ERROR_MOVE_STATE_OK = 0
const ERROR_MOVE_STATE_FAILED = 8
const ERROR_CREATE_OR_UPDATE_STATE_OK = 0
const ERROR_CREATE_OR_UPDATE_STATE_FAILED = 9
const ERROR_CAN_NOT_MOVE_KVSTATE = 10
const ERROR_CAN_NOT_MOVE_SETSTATE = 11
const ERROR_CREATE_SNAPSHOT_MIDIFY_HISTORY_FAILED = 12
const ERROR_UPDATE_SNAPSHOT_MIDIFY_HISTORY_FAILED = 13
const ERROR_FORCE_REFRESH_WINDOW = 14
const ERROR_MESSAGE_END = 15
const ERROR_CONTEXT_LENGTH_EXCEEDED = 16
const ERROR_INSUFFICIENT_COLLA = 17
const ERROR_AI_AGENT_MAX_TOKEN_OVER_COLLA_BALANCE = 18
const ERROR_FLAG_ONLY_PAID_TEAM_CAN_RUN_SPECIAL_AI_AGENT_MODEL = 19
const ERROR_PUT_STATE_FAILED = 20

type Feedback struct {
	ErrorCode    int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	Broadcast    *Broadcast  `json:"broadcast"`
	Data         interface{} `json:"data"`
}

func (feed *Feedback) Serialization() ([]byte, error) {
	return json.Marshal(feed)
}
