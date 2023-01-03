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

func SignalMoveState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
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
			currentNode.ConstructByMap(v) // set Name
			currentNode.ConstructByApp(appDto)
			currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS)

			if err := hub.TreeStateServiceImpl.MoveTreeStateNode(currentNode); err != nil {
				currentClient.Feedback(message, ws.ERROR_MOVE_STATE_FAILED, err)
				return err
			}
		}

	case ws.TARGET_DEPENDENCIES:
		fallthrough
	case ws.TARGET_DRAG_SHADOW:
		fallthrough
	case ws.TARGET_DOTTED_LINE_SQUARE:
		err := errors.New("K-V State do not support move method.")
		currentClient.Feedback(message, ws.ERROR_CAN_NOT_MOVE_KVSTATE, err)
		return nil
	case ws.TARGET_DISPLAY_NAME:
		err := errors.New("Set State do not support move method.")
		currentClient.Feedback(message, ws.ERROR_CAN_NOT_MOVE_SETSTATE, err)
		return nil
	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// the currentClient does not need feedback when operation success

	// change app modify time
	hub.AppServiceImpl.UpdateAppModifyTime(appDto)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)

	return nil
}
