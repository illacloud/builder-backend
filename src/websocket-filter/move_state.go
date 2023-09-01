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
	"github.com/illacloud/builder-backend/src/utils/builderoperation"
	"github.com/illacloud/builder-backend/src/websocket"
)

func SignalMoveState(hub *websocket.Hub, message *websocket.Message) error {
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
		currentClient.Feedback(message, websocket.ERROR_MOVE_STATE_FAILED, errInRetrieveApp)
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
			// init current tree state node
			currentTreeStateNode, errInInitCurrentNode := model.NewTreeStateByMoveStateWebsocketMessage(app, stateType, v)
			if errInInitCurrentNode != nil {
				currentClient.Feedback(message, websocket.ERROR_MOVE_STATE_FAILED, errInInitCurrentNode)
				return errInInitCurrentNode
			}

			errInMoveTreeState := hub.Storage.TreeStateStorage.MoveTreeStateNode(currentTreeStateNode)
			if errInMoveTreeState != nil {
				currentClient.Feedback(message, websocket.ERROR_MOVE_STATE_FAILED, errInMoveTreeState)
				return errInMoveTreeState
			}
			// collect display names
			displayNames = append(displayNames, currentTreeStateNode.ExportName())
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_DEPENDENCIES:
		fallthrough
	case builderoperation.TARGET_DRAG_SHADOW:
		fallthrough
	case builderoperation.TARGET_DOTTED_LINE_SQUARE:
		err := errors.New("K-V State do not support move method.")
		currentClient.Feedback(message, websocket.ERROR_CAN_NOT_MOVE_KVSTATE, err)
		return nil
	case builderoperation.TARGET_DISPLAY_NAME:
		err := errors.New("Set State do not support move method.")
		currentClient.Feedback(message, websocket.ERROR_CAN_NOT_MOVE_SETSTATE, err)
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
