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
	"errors"

	"github.com/illacloud/builder-backend/internal/repository"
	ws "github.com/illacloud/builder-backend/internal/websocket"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"
)

func SignalUpdateState(hub *ws.Hub, message *ws.Message) error {

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
		stateType = repository.TREE_STATE_TYPE_COMPONENTS
		for _, v := range message.Payload {
			// construct payload
			csfu, err := repository.ConstructComponentStateForUpdateByPayload(v)
			if err != nil {
				return err
			}

			// find component id by displayName
			beforeTreeState := state.NewTreeStateDto()
			beforeTreeState.ConstructByApp(appDto)
			beforeTreeState.ConstructWithType(stateType)
			beforeTreeState.ConstructByMap(csfu.Before)

			inDBTreeStateDto, err := hub.TreeStateServiceImpl.GetTreeStateByName(beforeTreeState)
			if err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_OR_UPDATE_STATE_FAILED, err)
				return err
			}

			if inDBTreeStateDto == nil {
				err := errors.New("[websocket-server] target state not exists, can not update.")
				currentClient.Feedback(message, ws.ERROR_CREATE_OR_UPDATE_STATE_FAILED, err)
				return nil
			}

			// construct update data
			component := repository.ConstructComponentNodeByMap(csfu.After)
			afterTreeState, err := hub.TreeStateServiceImpl.NewTreeStateByComponentState(appDto, component)
			if err != nil {
				return err
			}

			// update
			inDBTreeStateDto.ConstructWithNewStateContent(afterTreeState)
			if _, err := hub.TreeStateServiceImpl.UpdateTreeState(inDBTreeStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		// update k-v state
		for _, v := range message.Payload {
			subv, ok := v.(map[string]interface{})
			if !ok {
				err := errors.New("K-V State reflect failed, please check your input.")
				return err
			}
			for key, depState := range subv {
				// fill KVStateDto
				kvStateDto := state.NewKVStateDto()
				kvStateDto.ConstructWithKey(key)
				kvStateDto.ConstructForDependenciesState(depState)
				kvStateDto.ConstructByApp(appDto) // set AppRefID
				kvStateDto.ConstructWithType(stateType)

				if err := hub.KVStateServiceImpl.UpdateKVStateByKey(kvStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
					return err
				}
			}
		}

	case ws.TARGET_DRAG_SHADOW:
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
		// update K-V State
		for _, v := range message.Payload {
			// fill KVStateDto
			kvStateDto := state.NewKVStateDto()
			kvStateDto.ConstructByMap(v)
			kvStateDto.ConstructByApp(appDto)
			kvStateDto.ConstructWithType(stateType)

			// update
			if err := hub.KVStateServiceImpl.UpdateKVStateByKey(kvStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DISPLAY_NAME:
		stateType = repository.SET_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			// resolve payload
			dnsfu, err := repository.ConstructDisplayNameStateForUpdateByPayload(v)

			if err != nil {
				return err
			}
			// init state dto
			beforeSetStateDto := state.NewSetStateDto()
			afterSetStateDto := state.NewSetStateDto()
			beforeSetStateDto.ConstructWithValueBeforeUpdate(dnsfu)
			beforeSetStateDto.ConstructWithType(stateType)
			beforeSetStateDto.ConstructByApp(appDto)
			beforeSetStateDto.ConstructWithEditVersion()
			afterSetStateDto.ConstructWithValueAfterUpdate(dnsfu)

			// update state
			if err := hub.SetStateServiceImpl.UpdateSetStateByValue(beforeSetStateDto, afterSetStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_UPDATE_STATE_FAILED, err)
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
