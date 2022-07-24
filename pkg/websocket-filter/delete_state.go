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

	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalDeleteState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.APPID)
	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case ws.TARGET_NOTNING:
		return nil
	case ws.TARGET_COMPONENTS:
		for _, v := range message.Payload {
			var currentNode *state.TreeStateDto
			currentNode.ConstructByMap(v)      // set Name
			currentNode.ConstructByApp(appDto) // set AppRefID
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS)

			if err := hub.TreeStateServiceImpl.DeleteTreeStateNodeRecursive(currentNode); err != nil {
				currentClient.Feedback(message, ws.ws.ERROR_DELETE_STATE_FAILED, err)
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
		// delete k-v state
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvStateDto *state.KVStateDto
			kvStateDto.ConstructByMap(v)
			kvStateDto.ConstructByApp(appDto) // set AppRefID
			kvStateDto.ConstructWithType(stateType)

			if err := hub.KVStateServiceImpl.DeleteKVStateByKey(kvStateDto); err != nil {
				currentClient.Feedback(message, ws.ws.ERROR_DELETE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		// create dnsplayName state

		for _, v := range message.Payload {
			// resolve payload
			dns, err := repository.ConstructDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// save state
			for _, displayName := range dns {
				// init
				var setStateDto *state.SetStateDto
				setStateDto.ConstructWithValue(displayName)
				setStateDto.ConstructWithType(stateType)
				setStateDto.ConstructByApp(appDto)
				setStateDto.ConstructWithEditVersion()
				// delete state
				if err := hub.SetStateServiceImpl.DeleteSetStateByValue(setStateDto); err != nil {
					currentClient.Feedback(message, ws.ws.ERROR_CREATE_STATE_FAILED, err)
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
	currentClient.Feedback(message, ws.ws.ERROR_DELETE_STATE_OK, nil)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)
	return nil
}
