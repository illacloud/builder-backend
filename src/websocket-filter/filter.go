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
	"log"

	proto "github.com/golang/protobuf/proto"
	"github.com/illacloud/builder-backend/src/utils/builderoperation"
	"github.com/illacloud/builder-backend/src/websocket"
)

func Run(hub *websocket.Hub) {
	for {
		select {
		// handle register event
		case client := <-hub.Register:
			hub.Clients[client.ID] = client
		case client := <-hub.RegisterBinary:
			hub.BinaryClients[client.ID] = client
		// handle unregister events
		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client.ID]; ok {
				delete(hub.Clients, client.ID)
				close(client.Send)
			}
		// handle all hub broadcast events
		case message := <-hub.Broadcast:
			for _, client := range hub.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(hub.Clients, client.ID)
				}
			}
		// handle client on message event
		case message := <-hub.OnTextMessage:
			SignalFilter(hub, message)
		case message := <-hub.OnBinaryMessage:
			BinarySignalFilter(hub, message)
		}

	}
}

func SignalFilter(hub *websocket.Hub, message *websocket.Message) error {
	switch message.Signal {
	case builderoperation.SIGNAL_PING:
		return SignalPing(hub, message)
	case builderoperation.SIGNAL_ENTER:
		return SignalEnter(hub, message)
	case builderoperation.SIGNAL_LEAVE:
		return SignalLeave(hub, message)
	case builderoperation.SIGNAL_CREATE_STATE:
		return SignalCreateState(hub, message)
	case builderoperation.SIGNAL_DELETE_STATE:
		return SignalDeleteState(hub, message)
	case builderoperation.SIGNAL_UPDATE_STATE:
		return SignalUpdateState(hub, message)
	case builderoperation.SIGNAL_MOVE_STATE:
		return SignalMoveState(hub, message)
	case builderoperation.SIGNAL_CREATE_OR_UPDATE_STATE:
		return SignalCreateOrUpdateState(hub, message)
	case builderoperation.SIGNAL_BROADCAST_ONLY:
		return SignalBroadcastOnly(hub, message)
	case builderoperation.SIGNAL_PUT_STATE:
		return SignalPutState(hub, message)
	case builderoperation.SIGNAL_GLOBAL_BROADCAST_ONLY:
		return SignalGlobalBroadcastOnly(hub, message)
	case builderoperation.SIGNAL_COOPERATE_ATTACH:
		return SignalCooperateAttach(hub, message)
	case builderoperation.SIGNAL_COOPERATE_DISATTACH:
		return SignalCooperateDisattach(hub, message)
	default:
		return nil
	}
}

func BinarySignalFilter(hub *websocket.Hub, message []byte) error {
	binaryMessageType, errInGetMessageType := websocket.GetBinaryMessageType(message)
	if errInGetMessageType != nil {
		return errInGetMessageType
	}

	switch binaryMessageType {
	case websocket.BINARY_MESSAGE_TYPE_MOVING:
		// decode binary message
		movingMessageBin := &websocket.MovingMessageBin{}
		if errInParse := proto.Unmarshal(message, movingMessageBin); errInParse != nil {
			log.Printf("[BinarySignalFilter] Failed to parse message MovingMessageBin: ", errInParse)
			return errInParse
		}

		// process message
		MovingMessageFilter(hub, movingMessageBin)

	}
	return nil
}

func MovingMessageFilter(hub *websocket.Hub, message *websocket.MovingMessageBin) error {
	switch message.Signal {
	case builderoperation.SIGNAL_MOVE_STATE:
		return SignalMoveStateBinary(hub, message)
	case builderoperation.SIGNAL_MOVE_CURSOR:
		return SignalMoveCursorBinary(hub, message)
	default:
		return nil
	}
}

func OptionFilter(hub *websocket.Hub, client *websocket.Client, message *websocket.Message) error {
	return nil
}
