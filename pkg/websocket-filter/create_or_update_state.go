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
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	ws "github.com/illacloud/builder-backend/internal/websocket"
)

func SignalCreateOrUpdateState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	appDto := app.NewAppDto()
	appDto.ConstructWithID(currentClient.APPID)
	appDto.ConstructWithUpdateBy(currentClient.MappedUserID)
	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case ws.TARGET_NOTNING:
		return nil

	case ws.TARGET_COMPONENTS:
		for _, v := range message.Payload {
			// construct TreeStateDto
			currentNode := state.NewTreeStateDto()
			var inDBTreeStateDto *state.TreeStateDto
			currentNode.ConstructByMap(v)                                        // set Name
			currentNode.ConstructByApp(appDto)                                   // set AppRefID
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS) // set StateType

			// check if state already in database
			inDBTreeStateDto, _ = hub.TreeStateServiceImpl.GetTreeStateByName(currentNode)
			if inDBTreeStateDto == nil {
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
				inDBTreeStateDto.ConstructWithNewStateContent(currentNode)
				if _, err := hub.TreeStateServiceImpl.UpdateTreeState(inDBTreeStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
					return err
				}
			}
		}

	case ws.TARGET_DEPENDENCIES:
		// dependencies can not create or update by this method

	case ws.TARGET_DRAG_SHADOW:
		// create by displayName
		fallthrough

	case ws.TARGET_DOTTED_LINE_SQUARE:
		// fill type
		if message.Target == ws.TARGET_DEPENDENCIES {
			stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		} else if message.Target == ws.TARGET_DRAG_SHADOW {
			stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		} else {
			stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		}
		// resolve
		for _, v := range message.Payload {
			// construct KVStateDto
			kvStateDto := state.NewKVStateDto()
			var inDBkvStateDto *state.KVStateDto
			kvStateDto.ConstructByMap(v)
			kvStateDto.ConstructByApp(appDto)
			kvStateDto.ConstructWithType(stateType)

			inDBkvStateDto, _ = hub.KVStateServiceImpl.GetKVStateByKey(kvStateDto)
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
		stateType = repository.SET_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			var err error
			var displayName string
			// resolve payload
			displayName, err = repository.ResolveDisplayNameByPayload(v)

			if err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_OR_UPDATE_STATE_FAILED, err)
				return err
			}
			// create or update state

			// checkout
			setStateDto := state.NewSetStateDto()
			var setStateDtoInDB *state.SetStateDto
			setStateDto.ConstructWithValue(displayName)
			setStateDto.ConstructWithType(stateType)
			setStateDto.ConstructByApp(appDto)
			setStateDto.ConstructWithEditVersion()
			// lookup state
			setStateDtoInDB, _ = hub.SetStateServiceImpl.GetByValue(setStateDto)
			if setStateDtoInDB == nil {
				// create

				if _, err = hub.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
					return err
				}
			} else {
				// update
				setStateDtoInDB.ConstructWithValue(setStateDto.Value)
				if _, err = hub.SetStateServiceImpl.UpdateSetState(setStateDtoInDB); err != nil {
					currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
					return err
				}
			}
		}
	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_ACTION:
		// serve on HTTP API, this signal only for broadcast
	}

	// the currentClient does not need feedback when operation success

	// change app modify time
	hub.AppServiceImpl.UpdateAppModifyTime(appDto)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)
	return nil
}
