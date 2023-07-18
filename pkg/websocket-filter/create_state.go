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
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	"github.com/illacloud/builder-backend/internal/util/builderoperation"
	ws "github.com/illacloud/builder-backend/internal/websocket"
)

func SignalCreateState(hub *ws.Hub, message *ws.Message) error {
	// deserialize message
	currentClient, hit := hub.Clients[message.ClientID]
	if !hit {
		return errors.New("[SignalCreateState] target client(" + message.ClientID.String() + ") does dot exists.")
	}
	stateType := repository.STATE_TYPE_INVALIED
	teamID := currentClient.TeamID
	appID := currentClient.APPID
	userID := currentClient.MappedUserID
	appDto := app.NewAppDto()
	appDto.ConstructWithID(appID)
	appDto.ConstructWithUpdateBy(userID)
	appDto.SetTeamID(currentClient.TeamID)
	app := repository.NewAppWithID(currentClient.APPID, currentClient.TeamID, currentClient.MappedUserID)
	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case builderoperation.TARGET_NOTNING:
		return nil
	case builderoperation.TARGET_COMPONENTS:
		// build component tree from json
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			componentTree := repository.ConstructComponentNodeByMap(v)
			if err := hub.TreeStateServiceImpl.CreateComponentTree(app, 0, componentTree); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
			// collect display names
			repository.ExportComponentTreeAllDisplayNames(componentTree, displayNames)
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		displayNames := make([]string, 0)
		// create k-v state
		for _, v := range message.Payload {
			subv, ok := v.(map[string]interface{})
			if !ok {
				err := errors.New("K-V State reflect failed, please check your input.")
				return err
			}
			for key, depState := range subv {
				displayNames = append(displayNames, key)
				// fill KVStateDto
				kvStateDto := state.NewKVStateDto()
				kvStateDto.InitUID()
				kvStateDto.SetTeamID(teamID)
				kvStateDto.ConstructWithKey(key)
				kvStateDto.ConstructForDependenciesState(depState)
				kvStateDto.ConstructByApp(appDto) // set AppRefID
				kvStateDto.ConstructWithType(stateType)

				if _, err := hub.KVStateServiceImpl.CreateKVState(kvStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
					return err
				}
			}
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_DRAG_SHADOW:
		fallthrough

	case builderoperation.TARGET_DOTTED_LINE_SQUARE:
		// fill type
		if message.Target == builderoperation.TARGET_DRAG_SHADOW {
			stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		} else {
			stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		}
		// create k-v state
		for _, v := range message.Payload {
			// fill KVStateDto
			kvStateDto := state.NewKVStateDto()
			kvStateDto.InitUID()
			kvStateDto.SetTeamID(teamID)
			kvStateDto.ConstructByMap(v)
			kvStateDto.ConstructByApp(appDto) // set AppRefID
			kvStateDto.ConstructWithType(stateType)

			if _, err := hub.KVStateServiceImpl.CreateKVState(kvStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
		}

	case builderoperation.TARGET_DISPLAY_NAME:
		stateType = repository.SET_STATE_TYPE_DISPLAY_NAME
		displayNames := make([]string, 0)
		// create set state
		for _, v := range message.Payload {
			// resolve payload
			displayName, err := repository.ResolveDisplayNameByPayload(v)
			if err != nil {
				return err
			}
			// save state
			setStateDto := state.NewSetStateDto()
			setStateDto.InitUID()
			setStateDto.SetTeamID(teamID)
			setStateDto.ConstructByApp(appDto) // set AppRefID
			setStateDto.ConstructWithValue(displayName)
			setStateDto.ConstructWithType(stateType)
			// create state
			if _, err := hub.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
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
