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

// message protocol from client in json:
//
// {
//     "signal":number,
//     "option":number(work as int32 bit),
//     "payload":string
// }

const SIGNAL_LOGIN = 1
const SIGNAL_LOGOUT = 2
const SIGNAL_CREATE = 3
const SIGNAL_UPDATE = 4
const SIGNAL_DELETE = 5

const OPTION_BROADCAST_ROOM = 1 // 00000000000000000000000000000001; // use as signed int32 in typescript

type Message struct {
	RoomID      string
	Protocol    *Protocol
	RawProtocol []byte
}

type Protocol struct {
	Signal    int    `json:"signal"`
	Option    int    `json:"option"`
	StateName string `json:"stateName"`
	Payload   string `json:"payload"`
}

func NewMessage(roomID string, rawProtocol []byte) *Message {
	// init Action
	var protocol Protocol
	json.Unmarshal(rawProtocol, &protocol)
	return &Message{
		RoomID:      roomID,
		Protocol:    &protocol,
		RawProtocol: rawProtocol,
	}

}
