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

func SignalCreateOrUpdateState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto app.AppDto
	appDto.ConstructByWebSocketClient(currentClient)
	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case ws.TARGET_NOTNING:
		return nil

	case ws.TARGET_COMPONENTS:
		for _, v := range message.Payload {
			// construct TreeStateDto
			var err error
			var currentNode *state.TreeStateDto
			var inDBTreeState *state.TreeStateDto
			currentNode.ConstructByMap(v)                                        // set Name
			currentNode.ConstructByApp(appDto)                                   // set AppRefID
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS) // set StateType

			// check if state already in database
			inDBTreeState, err = hub.TreeStateServiceImpl.GetTreeStateByName(currentNode)
			if err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_OR_UPDATE_STATE_FAILED, err)
				return err
			}
			if inDBTreeState == nil {
				// current state did not in database, create
				var componentTree *repository.ComponentNode
				componentTree = repository.ConstructComponentNodeByMap(v)

				if err := hub.TreeStateServiceImpl.CreateComponentTree(appDto, 0, componentTree); err != nil {
					currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
					return err
				}
			} else {
				// hit, update it
				// construct update data
				componentNode := repository.ConstructComponentNodeByMap(v)
				serializedComponent, err := componentNode.SerializationForDatabase()
				if err != nil {
					currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
					return err
				}
				currentNode.ConstructWithContent(serializedComponent)
				currentNode.ConstructWithID(inDBTreeState.ID) // update by id

				// update
				if _, err := hub.TreeStateServiceImpl.UpdateTreeState(currentNode); err != nil {
					currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
					return err
				}
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
		for _, v := range message.Payload {
			// construct KVStateDto
			var err error
			var kvStateDto *state.KVStateDto
			var inDBkvStateDto *state.KVStateDto
			kvStateDto.ConstructByMap(v)
			kvStateDto.ConstructByApp(appDto)
			kvStateDto.ConstructWithType(stateType)

			inDBkvStateDto, err = hub.KVStateServiceImpl.GetKVStatesByKey(kvStateDto)
			if err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
			if inDBkvStateDto == nil {
				// current state did not in database, create
				if _, err := hub.KVStateServiceImpl.CreateKVState(kvStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
					return err
				}
			} else {
				// hit, update it
				kvStateDto.ConstructWithID(inDBkvStateDto.ID)
				if err := hub.KVStateServiceImpl.UpdateKVStateByID(kvStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
					return err
				}
			}

		}

	case ws.TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			// resolve payload
			displayNameState, err := repository.ResolveDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// create or update state
			for _, displayName := range displayNameState {
				// checkout
				var setStateDto state.SetStateDto
				var setStateDtoInDB *state.SetStateDto
				var err error
				setStateDto.ConstructWithValue(displayName)
				setStateDto.ConstructWithType(stateType)
				setStateDto.ConstructByApp(appDto)
				setStateDto.ConstructWithEditVersion()
				// lookup state
				if setStateDtoInDB, err = hub.SetStateServiceImpl.GetByValue(setStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
					return err
				}
				if setStateDtoInDB == nil {
					// create
					if _, err = hu.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
						currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
						return err
					}
				} else {
					// update
					setStateDtoInDB.ConstructWithValue(setStateDto.Value)
					if _, err = hu.SetStateServiceImpl.UpdateSetState(setStateDtoInDB); err != nil {
						currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
						return err
					}
				}
			}
		}
	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	currentClient.Feedback(message, ws.ERROR_CREATE_OR_UPDATE_STATE_OK, nil)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)
	return nil
}
