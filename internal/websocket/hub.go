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
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/illacloud/builder-backend/pkg/state"
)

// clients hub, maintains active clients and broadcast messags.
type Hub struct {
	// registered clients map
	Clients       map[uuid.UUID]*Client
	BinaryClients map[uuid.UUID]*Client

	// inbound messages from the clients.
	// try ```hub.Broadcast <- []byte(message)```
	Broadcast chan []byte

	// on message process
	OnTextMessage chan *Message

	OnBinaryMessage chan []byte

	// register requests from the clients.
	Register       chan *Client
	RegisterBinary chan *Client

	// unregister requests from the clients.
	Unregister chan *Client

	// InRoomUsers
	InRoomUsersMap map[int]*InRoomUsers // map[roomID]*InRoomUsers

	// impl
	TreeStateServiceImpl      *state.TreeStateServiceImpl
	KVStateServiceImpl        *state.KVStateServiceImpl
	SetStateServiceImpl       *state.SetStateServiceImpl
	AppServiceImpl            *app.AppServiceImpl
	ResourceServiceImpl       *resource.ResourceServiceImpl
	TreeStateRepositoryImpl   *repository.TreeStateRepositoryImpl
	KVStateRepositoryImpl     *repository.KVStateRepositoryImpl
	SetStateRepositoryImpl    *repository.SetStateRepositoryImpl
	ActionRepositoryImpl      *repository.ActionRepositoryImpl
	AppRepositoryImpl         *repository.AppRepositoryImpl
	AppSnapshotRepositoryImpl *repository.AppSnapshotRepositoryImpl
}

func NewHub() *Hub {
	return &Hub{
		Clients:         make(map[uuid.UUID]*Client),
		BinaryClients:   make(map[uuid.UUID]*Client),
		Broadcast:       make(chan []byte),
		OnTextMessage:   make(chan *Message),
		OnBinaryMessage: make(chan []byte),
		Register:        make(chan *Client),
		RegisterBinary:  make(chan *Client),
		Unregister:      make(chan *Client),
		InRoomUsersMap:  make(map[int]*InRoomUsers),
	}
}

func (hub *Hub) SetTreeStateServiceImpl(tssi *state.TreeStateServiceImpl) {
	hub.TreeStateServiceImpl = tssi
}

func (hub *Hub) SetKVStateServiceImpl(kvssi *state.KVStateServiceImpl) {
	hub.KVStateServiceImpl = kvssi
}

func (hub *Hub) SetSetStateServiceImpl(sssi *state.SetStateServiceImpl) {
	hub.SetStateServiceImpl = sssi
}

func (hub *Hub) SetAppServiceImpl(asi *app.AppServiceImpl) {
	hub.AppServiceImpl = asi
}

func (hub *Hub) SetResourceServiceImpl(rsi *resource.ResourceServiceImpl) {
	hub.ResourceServiceImpl = rsi
}

func (hub *Hub) SetAppRepositoryImpl(appRepositoryImpl *repository.AppRepositoryImpl) {
	hub.AppRepositoryImpl = appRepositoryImpl
}

func (hub *Hub) SetTreeStateRepositoryImpl(treeStateRepositoryImpl *repository.TreeStateRepositoryImpl) {
	hub.TreeStateRepositoryImpl = treeStateRepositoryImpl
}

func (hub *Hub) SetKVStateRepositoryImpl(kVStateRepositoryImpl *repository.KVStateRepositoryImpl) {
	hub.KVStateRepositoryImpl = kVStateRepositoryImpl
}

func (hub *Hub) SetSetStateRepositoryImpl(setStateRepositoryImpl *repository.SetStateRepositoryImpl) {
	hub.SetStateRepositoryImpl = setStateRepositoryImpl
}

func (hub *Hub) SetActionRepositoryImpl(actionRepositoryImpl *repository.ActionRepositoryImpl) {
	hub.ActionRepositoryImpl = actionRepositoryImpl
}

func (hub *Hub) SetAppSnapshotRepositoryImpl(appSnapshotRepositoryImpl *repository.AppSnapshotRepositoryImpl) {
	hub.AppSnapshotRepositoryImpl = appSnapshotRepositoryImpl
}

func (hub *Hub) GetInRoomUsersByRoomID(roomID int) *InRoomUsers {
	inRoomUsers, hit := hub.InRoomUsersMap[roomID]
	if !hit {
		hub.InRoomUsersMap[roomID] = NewInRoomUsers(roomID)
		return hub.InRoomUsersMap[roomID]
	}
	return inRoomUsers
}

func (hub *Hub) CleanRoom(roomID int) {
	inRoomUsers, hit := hub.InRoomUsersMap[roomID]
	if inRoomUsers.Count() != 0 || !hit {
		return
	}
	delete(hub.InRoomUsersMap, roomID)
}

func (hub *Hub) RemoveClient(client *Client) {
	close(client.Send)
	delete(hub.Clients, client.GetID())
	delete(hub.BinaryClients, client.GetID())
}

func (hub *Hub) BroadcastToOtherClients(message *Message, currentClient *Client) {
	if !message.NeedBroadcast {
		return
	}
	feedOtherClient := Feedback{
		ErrorCode:    ERROR_CODE_BROADCAST,
		ErrorMessage: "",
		Broadcast:    message.Broadcast,
		Data:         nil,
	}
	feedbyte, _ := feedOtherClient.Serialization()

	for clientid, client := range hub.Clients {
		if client.IsDead() {
			hub.RemoveClient(client)
			continue
		}
		if clientid == currentClient.ID {
			continue
		}
		if client.TeamID != currentClient.TeamID {
			continue
		}
		if client.APPID != currentClient.APPID {
			continue
		}
		client.Send <- feedbyte
	}
}

func (hub *Hub) BroadcastBinaryToOtherClients(message []byte, currentClient *Client) {
	for clientid, client := range hub.BinaryClients {
		if client.IsDead() {
			hub.RemoveClient(client)
			continue
		}
		if clientid == currentClient.ID {
			continue
		}
		if client.TeamID != currentClient.TeamID {
			continue
		}
		if client.APPID != currentClient.APPID {
			continue
		}
		client.Send <- message
	}
}

func (hub *Hub) BroadcastToRoomAllClients(message *Message, currentClient *Client) {
	if !message.NeedBroadcast {
		return
	}
	feedOtherClient := Feedback{
		ErrorCode:    ERROR_CODE_BROADCAST,
		ErrorMessage: "",
		Broadcast:    message.Broadcast,
		Data:         nil,
	}
	feedbyte, _ := feedOtherClient.Serialization()

	for _, client := range hub.Clients {
		if client.IsDead() {
			hub.RemoveClient(client)
			continue
		}
		if client.TeamID != currentClient.TeamID {
			continue
		}
		if client.APPID != currentClient.APPID {
			continue
		}
		client.Send <- feedbyte
	}
}

func (hub *Hub) BroadcastToTeamAllClients(message *Message, currentClient *Client, includeCurrentClient bool) {
	feed := Feedback{
		ErrorCode:    ERROR_CODE_BROADCAST,
		ErrorMessage: "",
		Broadcast:    message.Broadcast,
		Data:         nil,
	}
	feedbyte, _ := feed.Serialization()
	for clientid, client := range hub.Clients {
		if client.IsDead() {
			hub.RemoveClient(client)
			continue
		}
		if client.TeamID != currentClient.TeamID {
			continue
		}
		if clientid == currentClient.ID && !includeCurrentClient {
			continue
		}
		client.Send <- feedbyte
	}
}

// WARRING: This method will broadcast to server all clients. Use it carefully.
func (hub *Hub) BroadcastToGlobal(message *Message, currentClient *Client, includeCurrentClient bool) {
	feed := Feedback{
		ErrorCode:    ERROR_CODE_BROADCAST,
		ErrorMessage: "",
		Broadcast:    message.Broadcast,
		Data:         nil,
	}
	feedbyte, _ := feed.Serialization()
	for clientid, client := range hub.Clients {
		if client.IsDead() {
			hub.RemoveClient(client)
			continue
		}
		if clientid == currentClient.ID && !includeCurrentClient {
			continue
		}
		client.Send <- feedbyte
	}
}

func (hub *Hub) KickClient(client *Client) {
	close(client.Send)
	delete(hub.Clients, client.ID)
	delete(hub.BinaryClients, client.ID)
}
