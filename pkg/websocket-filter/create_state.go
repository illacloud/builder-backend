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

package filter

import (
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/state"

	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalCreateState(hub *ws.Hub, message *ws.Message) error {
	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto *app.AppDto
	appDto.ConstructWithID(currentClient.APPID)
	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case ws.TARGET_NOTNING:
		return nil
	case ws.TARGET_COMPONENTS:
		// build component tree from json
		for _, v := range message.Payload {
			componentTree := repository.ConstructComponentNodeByMap(v)
			if err := hub.TreeStateServiceImpl.CreateComponentTree(appDto, 0, componentTree); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough

	case ws.TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough

	case ws.TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// create k-v state
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvStateDto *state.KVStateDto
			kvStateDto.ConstructByMap(v)
			kvStateDto.ConstructWithType(stateType)

			if _, err := hub.KVStateServiceImpl.CreateKVState(kvStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		// create set state
		for _, v := range message.Payload {
			// resolve payload
			displayNameState, err := repository.ResolveDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// save state
			for _, displayName := range displayNameState {
				var setStateDto *state.SetStateDto
				setStateDto.ConstructWithValue(displayName)
				setStateDto.ConstructWithType(stateType)
				// create state
				if _, err := hub.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
					return err
				}
			}
		}

	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	currentClient.Feedback(message, ws.ERROR_CREATE_STATE_OK, nil)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)
	return nil
}
