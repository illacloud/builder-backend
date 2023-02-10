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
)

func SignalLeave(hub *ws.Hub, message *ws.Message) error {
	currentClient := hub.Clients[message.ClientID]

	// broadcast in room users
	inRoomUsers := hub.GetInRoomUsersByRoomID(currentClient.APPID)
	inRoomUsers.LeaveRoom(currentClient.MappedUserID)
	message.SetBroadcastType(ws.BROADCAST_TYPE_ENTER)
	message.RewriteBroadcast()
	message.SetBroadcastPayload(inRoomUsers.FetchAllInRoomUsers())
	hub.BroadcastToOtherClients(message, currentClient)

	
	// broadcast attached components users
	message.SetBroadcastType(ws.BROADCAST_TYPE_ATTACH_COMPONENT)
	message.RewriteBroadcast()
	message.SetBroadcastPayload(inRoomUsers.FetchAllAttachedUsers())
	hub.BroadcastToOtherClients(message, currentClient)

	// kick leaved user
	ws.KickClient(hub, currentClient)
	hub.CleanRoom(currentClient.APPID)
	return nil
}
