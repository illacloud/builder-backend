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

const SIGNAL_PING = 0
const SIGNAL_ENTER = 1
const SIGNAL_LEAVE = 2
const SIGNAL_CREATE_STATE = 3
const SIGNAL_DELETE_STATE = 4
const SIGNAL_UPDATE_STATE = 5
const SIGNAL_MOVE_STATE = 6
const SIGNAL_CREATE_OR_UPDATE = 7
const SIGNAL_ONLY_BROADCAST = 8

const OPTION_BROADCAST_ROOM = 1 // 00000000000000000000000000000001; // use as signed int32 in typescript

const TARGET_NOTNING = 0            // placeholder for nothing
const TARGET_COMPONENTS = 1         // ComponentsState
const TARGET_DEPENDENCIES = 2       // DependenciesState
const TARGET_DRAG_SHADOW = 3        // DragShadowState
const TARGET_DOTTED_LINE_SQUARE = 4 // DottedLineSquareState
const TARGET_DISPLAY_NAME = 5       // DisplayNameState
const TARGET_APPS = 6               // only for broadcast
const TARGET_RESOURCE = 7           // only for broadcast

type Broadcast struct {
    Type    string `json:"type"`
    Payload string `json:"payload"`
}

type Protocol struct {
    Signal    int                    `json:"signal"`
    Option    int                    `json:"option"`
    Target    int                    `json:"target"`
    Payload   map[string]interface{} `json:"payload"`
    Broadcast *Broadcast             `json:"broadcast"`
}

func NewProtocol(roomID string, rawProtocol []byte) (*Protocol, error) {
    // init Action
    var protocol Protocol
    if err := json.Unmarshal(rawProtocol, &protocol); err != nil {
        return nil, err
    }
    return &protocol, nil
}
