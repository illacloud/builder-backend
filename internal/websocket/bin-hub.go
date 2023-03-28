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

package ws

import (
	"github.com/google/uuid"
)

// clients hub, maintains active clients and broadcast messags.
type BinHub struct {
	// registered clients map
	Clients map[uuid.UUID]*Client

	// inbound messages from the clients.
	// try ```hub.Broadcast <- []byte(message)```
	Broadcast chan []byte

	// on message process
	OnBinaryMessage chan []byte

	// register requests from the clients.
	Register chan *Client

	// unregister requests from the clients.
	Unregister chan *Client

	// InRoomUsers
	InRoomUsersMap map[int]*InRoomUsers // map[roomID]*InRoomUsers
}

func NewBinHub() *BinHub {
	return &BinHub{
		Clients:         make(map[uuid.UUID]*Client),
		Broadcast:       make(chan []byte),
		OnBinaryMessage: make(chan []byte),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		InRoomUsersMap:  make(map[int]*InRoomUsers),
	}
}

func (hub *BinHub) GetInRoomUsersByRoomID(roomID int) *InRoomUsers {
	inRoomUsers, hit := hub.InRoomUsersMap[roomID]
	if !hit {
		hub.InRoomUsersMap[roomID] = NewInRoomUsers(roomID)
		return hub.InRoomUsersMap[roomID]
	}
	return inRoomUsers
}

func (hub *BinHub) CleanRoom(roomID int) {
	inRoomUsers, hit := hub.InRoomUsersMap[roomID]
	if inRoomUsers.Count() != 0 || !hit {
		return
	}
	delete(hub.InRoomUsersMap, roomID)
}

func (hub *BinHub) BroadcastToOtherClients(message []byte, currentClient *Client) {
	for clientid, client := range hub.Clients {
		if clientid == currentClient.ID {
			continue
		}
		if client.APPID != currentClient.APPID {
			continue
		}
		client.Send <- message
	}
}

func (hub *BinHub) BroadcastToRoomAllClients(message []byte, currentClient *Client) {
	for _, client := range hub.Clients {
		if client.APPID != currentClient.APPID {
			continue
		}
		client.Send <- message
	}
}

func (hub *BinHub) BroadcastToGlobal(message []byte, currentClient *Client, includeCurrentClient bool) {
	for clientid, client := range hub.Clients {
		if clientid == currentClient.ID && !includeCurrentClient {
			continue
		}
		client.Send <- message
	}
}

func (hub *BinHub) KickClient(client *Client) {
	close(client.Send)
	delete(hub.Clients, client.ID)
}
