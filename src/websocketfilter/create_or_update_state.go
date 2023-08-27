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
	"github.com/illacloud/builder-backend/internal/util/builderoperation"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	ws "github.com/illacloud/builder-backend/internal/websocket"
)

func SignalCreateOrUpdateState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient, hit := hub.Clients[message.ClientID]
	if !hit {
		return errors.New("[SignalCreateOrUpdateState] target client(" + message.ClientID.String() + ") does dot exists.")
	}
	stateType := repository.STATE_TYPE_INVALIED
	teamID := currentClient.TeamID
	appID := currentClient.APPID
	userID := currentClient.MappedUserID

	appDto := app.NewAppDto()
	appDto.InitUID()
	appDto.ConstructWithID(appID)
	appDto.SetTeamID(currentClient.TeamID)
	appDto.ConstructWithUpdateBy(userID)
	message.RewriteBroadcast()
	app := repository.NewApp("", teamID, userID)

	// target switch
	switch message.Target {
	case builderoperation.TARGET_NOTNING:
		return nil

	case builderoperation.TARGET_COMPONENTS:
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			// construct TreeStateDto
			var inDBTreeStateDto *state.TreeStateDto
			currentNode := state.NewTreeStateDto()
			// @todo: refactor this to fix arity new function.
			currentNode.InitUID()                                                // set UID
			currentNode.SetTeamID(teamID)                                        // set teamID
			currentNode.ConstructByMap(v)                                        // set Name
			currentNode.ConstructByApp(appDto)                                   // set AppRefID
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS) // set StateType

			// check if state already in database
			inDBTreeStateDto, _ = hub.TreeStateServiceImpl.GetTreeStateByName(currentNode)
			if inDBTreeStateDto == nil {
				// current state did not in database, create
				var componentTree *repository.ComponentNode
				componentTree = repository.ConstructComponentNodeByMap(v)

				if err := hub.TreeStateServiceImpl.CreateComponentTree(app, 0, componentTree); err != nil {
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
			displayNames = append(displayNames, currentNode.ExportName())
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_DEPENDENCIES:
		// dependencies can not create or update by this method

	case builderoperation.TARGET_DRAG_SHADOW:
		// create by displayName
		fallthrough

	case builderoperation.TARGET_DOTTED_LINE_SQUARE:
		// fill type
		if message.Target == builderoperation.TARGET_DEPENDENCIES {
			stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		} else if message.Target == builderoperation.TARGET_DRAG_SHADOW {
			stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		} else {
			stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		}
		// resolve
		for _, v := range message.Payload {
			// construct KVStateDto
			var inDBkvStateDto *state.KVStateDto
			kvStateDto := state.NewKVStateDto()
			kvStateDto.InitUID()
			kvStateDto.SetTeamID(teamID)
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

	case builderoperation.TARGET_DISPLAY_NAME:
		stateType = repository.SET_STATE_TYPE_DISPLAY_NAME
		displayNames := make([]string, 0)
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
			var setStateDtoInDB *state.SetStateDto
			setStateDto := state.NewSetStateDto()
			setStateDto.InitUID()
			setStateDto.SetTeamID(teamID)
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
			displayNames = append(displayNames, displayName)
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			appForExport, errInNewAppForExport := repository.NewAppForExportByMap(v)
			if errInNewAppForExport == nil {
				displayNames = append(displayNames, appForExport.ExportName())
			}
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			resourceForExport, errInNewResourceForExport := repository.NewResourceForExportByMap(v)
			if errInNewResourceForExport == nil {
				displayNames = append(displayNames, resourceForExport.ExportName())
			}
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_ACTION:
		// serve on HTTP API, this signal only for broadcast
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			actionForExport, errInNewActionForExport := repository.NewActionForExportByMap(v)
			if errInNewActionForExport == nil {
				displayNames = append(displayNames, actionForExport.ExportDisplayName())
			}
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	}

	// the currentClient does not need feedback when operation success

	// change app modify time
	hub.AppServiceImpl.UpdateAppModifyTime(appDto)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)
	return nil
}
