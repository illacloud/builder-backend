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
	ws "github.com/illacloud/builder-backend/internal/websocket"
	"github.com/illacloud/builder-backend/pkg/user"
)

func Run(hub *ws.Hub) {
	for {
		select {
		// handle register event
		case client := <-hub.Register:
			hub.Clients[client.ID] = client
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
		case message := <-hub.OnMessage:
			SignalFilter(hub, message, hub.AuthenticatorImpl)
		}

	}
}

func SignalFilter(hub *ws.Hub, message *ws.Message, ai *user.AuthenticatorImpl) error {
	switch message.Signal {
	case ws.SIGNAL_PING:
		return SignalPing(hub, message)
	case ws.SIGNAL_ENTER:
		return SignalEnter(hub, message, ai)
	case ws.SIGNAL_LEAVE:
		return SignalLeave(hub, message)
	case ws.SIGNAL_CREATE_STATE:
		return SignalCreateState(hub, message)
	case ws.SIGNAL_DELETE_STATE:
		return SignalDeleteState(hub, message)
	case ws.SIGNAL_UPDATE_STATE:
		return SignalUpdateState(hub, message)
	case ws.SIGNAL_MOVE_STATE:
		return SignalMoveState(hub, message)
	case ws.SIGNAL_CREATE_OR_UPDATE_STATE:
		return SignalCreateOrUpdateState(hub, message)
	case ws.SIGNAL_BROADCAST_ONLY:
		return SignalBroadcastOnly(hub, message)
	case ws.SIGNAL_PUT_STATE:
		return SignalPutState(hub, message)
	case ws.SIGNAL_GLOBAL_BROADCAST_ONLY:
		return SignalGlobalBroadcastOnly(hub, message)
	case ws.SIGNAL_COOPERATE_ATTACH:
		return SignalCooperateAttach(hub, message)
	case ws.SIGNAL_COOPERATE_DISATTACH:
		return SignalCooperateDisattach(hub, message)
	default:
		return nil
	}
}

func OptionFilter(hub *ws.Hub, client *ws.Client, message *ws.Message) error {
	return nil
}
