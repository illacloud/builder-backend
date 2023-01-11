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

func SignalDeleteState(hub *ws.Hub, message *ws.Message) error {

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
			currentNode := state.NewTreeStateDto()
			currentNode.ConstructWithDisplayNameForDelete(v) // set Name
			currentNode.ConstructByApp(appDto)               // set AppRefID
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS)

			if err := hub.TreeStateServiceImpl.DeleteTreeStateNodeRecursive(currentNode); err != nil {
				currentClient.Feedback(message, ws.ERROR_DELETE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DEPENDENCIES:
		// dependency can not delete

	case ws.TARGET_DRAG_SHADOW:
		fallthrough

	case ws.TARGET_DOTTED_LINE_SQUARE:
		// fill type
		if message.Target == ws.TARGET_DRAG_SHADOW {
			stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		} else {
			stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		}
		// delete k-v state
		for _, v := range message.Payload {
			// fill KVStateDto
			kvStateDto := state.NewKVStateDto()
			kvStateDto.ConstructWithDisplayNameForDelete(v)
			kvStateDto.ConstructByApp(appDto) // set AppRefID
			kvStateDto.ConstructWithType(stateType)

			if err := hub.KVStateServiceImpl.DeleteKVStateByKey(kvStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_DELETE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DISPLAY_NAME:
		stateType = repository.SET_STATE_TYPE_DISPLAY_NAME
		// delete set state
		for _, v := range message.Payload {

			// init
			setStateDto := state.NewSetStateDto()
			setStateDto.ConstructWithDisplayNameForDelete(v)
			setStateDto.ConstructWithType(stateType)
			setStateDto.ConstructByApp(appDto)
			setStateDto.ConstructWithEditVersion()
			// delete state
			if err := hub.SetStateServiceImpl.DeleteSetStateByValue(setStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
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
