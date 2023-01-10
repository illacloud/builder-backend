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

package ws

import (
	"encoding/json"
	uuid "github.com/satori/go.uuid"
)

// message protocol from client in json:
//
// {
//     "signal":number,
//     "option":number(work as int32 bit),
//     "payload":string
// }

// for message
const SIGNAL_PING = 0
const SIGNAL_ENTER = 1
const SIGNAL_LEAVE = 2
const SIGNAL_CREATE_STATE = 3
const SIGNAL_DELETE_STATE = 4
const SIGNAL_UPDATE_STATE = 5
const SIGNAL_MOVE_STATE = 6
const SIGNAL_CREATE_OR_UPDATE_STATE = 7
const SIGNAL_BROADCAST_ONLY = 8
const SIGNAL_PUT_STATE = 9
const SIGNAL_GLOBAL_BROADCAST_ONLY = 10
const SIGNAL_COOPERATE_ATTACH = 11
const SIGNAL_COOPERATE_DISATTACH = 12

const OPTION_BROADCAST_ROOM = 1 // 00000000000000000000000000000001; // use as signed int32 in typescript

const TARGET_NOTNING = 0            // placeholder for nothing
const TARGET_COMPONENTS = 1         // ComponentsState
const TARGET_DEPENDENCIES = 2       // DependenciesState
const TARGET_DRAG_SHADOW = 3        // DragShadowState
const TARGET_DOTTED_LINE_SQUARE = 4 // DottedLineSquareState
const TARGET_DISPLAY_NAME = 5       // DisplayNameState
const TARGET_APPS = 6               // only for broadcast
const TARGET_RESOURCE = 7           // only for broadcast
const TARGET_ACTION = 8             // only for broadcast

// for broadcast rewrite
const BROADCAST_TYPE_SUFFIX = "/remote"
const BROADCAST_TYPE_ENTER = "enter"
const BROADCAST_TYPE_ATTACH_COMPONENT = "attachComponent"

type Broadcast struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Message struct {
	ClientID      uuid.UUID     `json:"clientID"`
	Signal        int           `json:"signal"`
	APPID         int           `json:"appID"` // also as APP ID
	Option        int           `json:"option"`
	Payload       []interface{} `json:"payload"`
	Target        int           `json:"target"`
	Broadcast     *Broadcast    `json:"broadcast"`
	NeedBroadcast bool
}

func NewMessage(clientID uuid.UUID, appID int, rawMessage []byte) (*Message, error) {
	// init Action
	var message Message
	if err := json.Unmarshal(rawMessage, &message); err != nil {
		return nil, err
	}
	message.ClientID = clientID
	message.APPID = appID
	if message.Broadcast == nil {
		message.NeedBroadcast = false
	} else {
		message.NeedBroadcast = true
	}
	return &message, nil
}

func (m *Message) SetSignal(s int) {
	m.Signal = SIGNAL_COOPERATE_ATTACH
}

func (m *Message) SetBroadcastType(t string) {
	if m.Broadcast != nil {
		m.Broadcast.Type = t
	}
}

func (m *Message) SetBroadcastPayload(any interface{}) {
	if m.Broadcast != nil {
		m.Broadcast.Payload = any
	}
}

func (m *Message) RewriteBroadcast() {
	if m.NeedBroadcast {
		m.Broadcast.Type = m.Broadcast.Type + BROADCAST_TYPE_SUFFIX
	}
}
