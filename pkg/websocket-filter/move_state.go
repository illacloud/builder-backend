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
	ws "github.com/illacloud/builder-backend/internal/websocket"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"
)

func SignalMoveState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient, hit := hub.Clients[message.ClientID]
	if !hit {
		return errors.New("[SignalMoveState] target client(" + message.ClientID.String() + ") does dot exists.")
	}
	teamID := currentClient.TeamID
	appDto := app.NewAppDto()
	appDto.ConstructWithID(currentClient.APPID)
	appDto.ConstructWithUpdateBy(currentClient.MappedUserID)
	appDto.SetTeamID(currentClient.TeamID)
	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case builderoperation.TARGET_NOTNING:
		return nil
	case builderoperation.TARGET_COMPONENTS:
		displayNames := make([]string, 0)
		for _, v := range message.Payload {
			currentNode := state.NewTreeStateDto()
			currentNode.InitUID()
			currentNode.SetTeamID(teamID)
			currentNode.ConstructByMap(v) // set Name
			currentNode.ConstructByApp(appDto)
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS)

			if err := hub.TreeStateServiceImpl.MoveTreeStateNode(currentNode); err != nil {
				currentClient.Feedback(message, ws.ERROR_MOVE_STATE_FAILED, err)
				return err
			}
			// collect display names
			displayNames = append(displayNames, currentNode.ExportName())
		}
		// record app snapshot modify history
		RecordModifyHistory(hub, message, displayNames)
	case builderoperation.TARGET_DEPENDENCIES:
		fallthrough
	case builderoperation.TARGET_DRAG_SHADOW:
		fallthrough
	case builderoperation.TARGET_DOTTED_LINE_SQUARE:
		err := errors.New("K-V State do not support move method.")
		currentClient.Feedback(message, ws.ERROR_CAN_NOT_MOVE_KVSTATE, err)
		return nil
	case builderoperation.TARGET_DISPLAY_NAME:
		err := errors.New("Set State do not support move method.")
		currentClient.Feedback(message, ws.ERROR_CAN_NOT_MOVE_SETSTATE, err)
		return nil
	case builderoperation.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case builderoperation.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// the currentClient does not need feedback when operation success

	// change app modify time
	hub.AppServiceImpl.UpdateAppModifyTime(appDto)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)

	return nil
}
