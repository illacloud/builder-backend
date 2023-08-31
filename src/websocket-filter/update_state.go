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
	"fmt"

	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/builderoperation"
	"github.com/illacloud/builder-backend/src/websocket"
)

func SignalUpdateState(hub *websocket.Hub, message *websocket.Message) error {
	// init global param
	currentClient, errInGetClient := hub.GetClientByID(message.ClientID)
	if errInGetClient != nil {
		return errInGetClient
	}
	stateType := model.STATE_TYPE_INVALIED
	teamID := currentClient.TeamID
	appID := currentClient.APPID
	userID := currentClient.MappedUserID

	// rewrite message broadcast
	message.RewriteBroadcast()

	// new app
	app, errInRetrieveApp := hub.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInRetrieveApp)
		return errInRetrieveApp
	}
	app.Modify(userID)

	// modified displayNames
	displayNames := make([]string, 0)

	// target switch
	switch message.Target {
	case builderoperation.TARGET_NOTNING:
		return nil
	case builderoperation.TARGET_COMPONENTS:
		stateType = model.TREE_STATE_TYPE_COMPONENTS
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			fmt.Printf("[DUMP] v: %+v\n", v)
			componentStateForUpdate, errInConstructComponentStateForUpdate := model.ConstructComponentStateForUpdateByPayload(v)
			if errInConstructComponentStateForUpdate != nil {
				currentClient.Feedback(message, websocket.ERROR_UPDATE_STATE_FAILED, errInConstructComponentStateForUpdate)
				return errInConstructComponentStateForUpdate
			}

			// init before state
			beforeTreeStateNode, errInInitCurrentNode := model.NewTreeStateByWebsocketMessage(app, stateType, componentStateForUpdate.Before)
			inDatabaseTreeState, errInRetrieveTreeState := hub.Storage.TreeStateStorage.RetrieveEditVersionByAppAndName(teamID, appID, beforeTreeStateNode.ExportStateType(), beforeTreeStateNode.ExportName())
			if errInRetrieveTreeState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInRetrieveTreeState)
				return errInRetrieveTreeState
			}
			fmt.Printf("[DUMP] inDatabaseTreeState: %+v\n", inDatabaseTreeState)

			// init current tree state node
			componentNodeAfter := model.ConstructComponentNodeByMap(componentStateForUpdate.After)
			currentTreeStateNode, errInInitCurrentNode := model.NewTreeStateByAppAndComponentState(app, stateType, componentNodeAfter)
			if errInInitCurrentNode != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInInitCurrentNode)
				return errInInitCurrentNode
			}
			fmt.Printf("[DUMP] currentTreeStateNode: %+v\n", currentTreeStateNode)

			// it is in database, update it
			// NOTE, tree_state with SIGNAL_UPDATE_STATE, will only update name and content field.
			inDatabaseTreeState.UpdateNameAndContent(currentTreeStateNode)
			fmt.Printf("[DUMP] inDatabaseTreeState.UpdateByNewTreeState: %+v\n", inDatabaseTreeState)

			errInUpdateTreeState := hub.Storage.TreeStateStorage.Update(inDatabaseTreeState)
			if errInUpdateTreeState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInUpdateTreeState)
				return errInUpdateTreeState
			}
			displayNames = append(displayNames, currentTreeStateNode.ExportName())
		}
	case builderoperation.TARGET_DEPENDENCIES:
		stateType = model.KV_STATE_TYPE_DEPENDENCIES
		displayNames := make([]string, 0)
		// update k-v state
		for _, v := range message.Payload {
			subv, ok := v.(map[string]interface{})
			if !ok {
				err := errors.New("K-V State reflect failed, please check your input.")
				return err
			}
			for key, depState := range subv {
				// init current kvState node
				currentKVStateNode, errInNewKVState := model.NewKVStateByWebsocketMessageWithGivenKey(app, stateType, key, depState)
				if errInNewKVState != nil {
					currentClient.Feedback(message, websocket.ERROR_CREATE_STATE_FAILED, errInNewKVState)
					return errInNewKVState
				}

				// check if state already in database
				inDatabaseKVState, errInRetrieveKVState := hub.Storage.KVStateStorage.RetrieveEditVersionByAppAndKey(teamID, appID, stateType, currentKVStateNode.ExportKey())
				if errInRetrieveKVState != nil {
					currentClient.Feedback(message, websocket.ERROR_CREATE_STATE_FAILED, errInRetrieveKVState)
					return errInRetrieveKVState
				}

				// hit, update it
				inDatabaseKVState.UpdateByNewKVState(currentKVStateNode)
				errInUpdateKVState := hub.Storage.KVStateStorage.Update(inDatabaseKVState)
				if errInUpdateKVState != nil {
					currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInUpdateKVState)
					return errInUpdateKVState
				}
				displayNames = append(displayNames, currentKVStateNode.ExportKey())
			}
		}
	case builderoperation.TARGET_DRAG_SHADOW:
		fallthrough

	case builderoperation.TARGET_DOTTED_LINE_SQUARE:
		// fill type
		if message.Target == builderoperation.TARGET_DEPENDENCIES {
			stateType = model.KV_STATE_TYPE_DEPENDENCIES
		} else if message.Target == builderoperation.TARGET_DRAG_SHADOW {
			stateType = model.KV_STATE_TYPE_DRAG_SHADOW
		} else {
			stateType = model.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		}
		// update K-V State
		for _, v := range message.Payload {
			// init current kvState node
			currentKVStateNode, errInNewKVState := model.NewKVStateByWebsocketMessage(app, stateType, v)
			if errInNewKVState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInNewKVState)
				return errInNewKVState
			}

			// check if state already in database
			inDatabaseKVState, errInRetrieveKVState := hub.Storage.KVStateStorage.RetrieveEditVersionByAppAndKey(teamID, appID, stateType, currentKVStateNode.ExportKey())
			if errInRetrieveKVState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_STATE_FAILED, errInRetrieveKVState)
				return errInRetrieveKVState
			}

			// hit, update it
			inDatabaseKVState.UpdateByNewKVState(currentKVStateNode)
			errInUpdateKVState := hub.Storage.KVStateStorage.Update(inDatabaseKVState)
			if errInUpdateKVState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInUpdateKVState)
				return errInUpdateKVState
			}
			displayNames = append(displayNames, currentKVStateNode.ExportKey())
		}

	case builderoperation.TARGET_DISPLAY_NAME:
		stateType = model.SET_STATE_TYPE_DISPLAY_NAME
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			// init current set state node
			currentSetStateNode, errInNewSetState := model.NewSetStateByWebsocketMessage(app, stateType, v)
			if errInNewSetState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInNewSetState)
				return errInNewSetState
			}
			// lookup state
			inDatabaseSetState, errInRetrieveSetState := hub.Storage.SetStateStorage.RetrieveByValue(currentSetStateNode)
			if errInRetrieveSetState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_STATE_FAILED, errInRetrieveSetState)
				return errInRetrieveSetState
			}
			// update
			inDatabaseSetState.UpdateByNewSetState(currentSetStateNode)
			errInUpdateSetState := hub.Storage.SetStateStorage.Update(inDatabaseSetState)
			if errInUpdateSetState != nil {
				currentClient.Feedback(message, websocket.ERROR_CREATE_OR_UPDATE_STATE_FAILED, errInUpdateSetState)
				return errInUpdateSetState
			}
			displayNames = append(displayNames, currentSetStateNode.ExportValue())
		}
	case builderoperation.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
		for _, v := range message.Payload {
			appForExport, errInNewAppForExport := model.NewAppForExportByMap(v)
			if errInNewAppForExport == nil {
				displayNames = append(displayNames, appForExport.ExportName())
			}
		}
	case builderoperation.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
		for _, v := range message.Payload {
			resourceForExport, errInNewResourceForExport := model.NewResourceForExportByMap(v)
			if errInNewResourceForExport == nil {
				displayNames = append(displayNames, resourceForExport.ExportName())
			}
		}
	case builderoperation.TARGET_ACTION:
		// serve on HTTP API, this signal only for broadcast
		for _, v := range message.Payload {
			actionForExport, errInNewActionForExport := model.NewActionForExportByMap(v)
			if errInNewActionForExport == nil {
				displayNames = append(displayNames, actionForExport.ExportDisplayName())
			}
		}
	}

	// record app snapshot modify history
	RecordModifyHistory(hub, message, displayNames)

	// the currentClient does not need feedback when operation success

	// change app modify time
	hub.Storage.AppStorage.UpdateWholeApp(app)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)

	return nil
}
