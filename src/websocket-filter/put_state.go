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

	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/websocket"

	"github.com/illacloud/builder-backend/src/utils/builderoperation"
)

func SignalPutState(hub *websocket.Hub, message *websocket.Message) error {
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
		currentClient.Feedback(message, websocket.ERROR_PUT_STATE_FAILED, errInRetrieveApp)
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
		return nil

	case builderoperation.TARGET_DEPENDENCIES:
		stateType = model.KV_STATE_TYPE_DEPENDENCIES
		errInDelteAllEditVersionKVState := hub.Storage.KVStateStorage.DeleteAllKVStatesByAppVersionAndType(teamID, appID, model.APP_EDIT_VERSION, stateType)
		if errInDelteAllEditVersionKVState != nil {
			return errInDelteAllEditVersionKVState
		}
		// create k-v state
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
					currentClient.Feedback(message, websocket.ERROR_PUT_STATE_FAILED, errInNewKVState)
					return errInNewKVState
				}

				// current state did not in database, create
				errInCreateKVState := hub.Storage.KVStateStorage.Create(currentKVStateNode)
				if errInCreateKVState != nil {
					currentClient.Feedback(message, websocket.ERROR_PUT_STATE_FAILED, errInCreateKVState)
					return errInCreateKVState
				}
			}
		}
	case builderoperation.TARGET_DRAG_SHADOW:
		return nil

	case builderoperation.TARGET_DOTTED_LINE_SQUARE:
		return nil

	case builderoperation.TARGET_DISPLAY_NAME:
		return nil

	case builderoperation.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case builderoperation.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
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
