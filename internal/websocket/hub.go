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
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/illacloud/builder-backend/pkg/state"
)

type Hub interface {
	Register chan *Client
	Unregister chan *Client
	KickClient(client *Client)
}

// clients hub, maintains active clients and broadcast messags.
type TextHub struct {
	// registered clients map
	Clients map[uuid.UUID]*Client

	// inbound messages from the clients.
	// try ```hub.Broadcast <- []byte(message)```
	Broadcast chan []byte

	// on message process
	OnMessage chan *Message

	// register requests from the clients.
	Register chan *Client

	// unregister requests from the clients.
	Unregister chan *Client

	// InRoomUsers
	InRoomUsersMap map[int]*InRoomUsers // map[roomID]*InRoomUsers

	// impl
	TreeStateServiceImpl *state.TreeStateServiceImpl
	KVStateServiceImpl   *state.KVStateServiceImpl
	SetStateServiceImpl  *state.SetStateServiceImpl
	AppServiceImpl       *app.AppServiceImpl
	ResourceServiceImpl  *resource.ResourceServiceImpl
}

func NewTextHub() *TextHub {
	return &TextHub{
		Clients:        make(map[uuid.UUID]*Client),
		Broadcast:      make(chan []byte),
		OnMessage:      make(chan *Message),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		InRoomUsersMap: make(map[int]*InRoomUsers),
	}
}

func (hub *TextHub) SetTreeStateServiceImpl(tssi *state.TreeStateServiceImpl) {
	hub.TreeStateServiceImpl = tssi
}

func (hub *TextHub) SetKVStateServiceImpl(kvssi *state.KVStateServiceImpl) {
	hub.KVStateServiceImpl = kvssi
}

func (hub *TextHub) SetSetStateServiceImpl(sssi *state.SetStateServiceImpl) {
	hub.SetStateServiceImpl = sssi
}

func (hub *TextHub) SetAppServiceImpl(asi *app.AppServiceImpl) {
	hub.AppServiceImpl = asi
}

func (hub *TextHub) SetResourceServiceImpl(rsi *resource.ResourceServiceImpl) {
	hub.ResourceServiceImpl = rsi
}

func (hub *TextHub) GetInRoomUsersByRoomID(roomID int) *InRoomUsers {
	inRoomUsers, hit := hub.InRoomUsersMap[roomID]
	if !hit {
		hub.InRoomUsersMap[roomID] = NewInRoomUsers(roomID)
		return hub.InRoomUsersMap[roomID]
	}
	return inRoomUsers
}

func (hub *TextHub) CleanRoom(roomID int) {
	inRoomUsers, hit := hub.InRoomUsersMap[roomID]
	if inRoomUsers.Count() != 0 || !hit {
		return
	}
	delete(hub.InRoomUsersMap, roomID)
}

func (hub *TextHub) BroadcastToOtherClients(message *Message, currentClient *Client) {
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
		if clientid == currentClient.ID {
			continue
		}
		if client.APPID != currentClient.APPID {
			continue
		}
		client.Send <- feedbyte
	}
}

func (hub *TextHub) BroadcastToRoomAllClients(message *Message, currentClient *Client) {
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
		if client.APPID != currentClient.APPID {
			continue
		}
		client.Send <- feedbyte
	}
}

func (hub *TextHub) BroadcastToGlobal(message *Message, currentClient *Client, includeCurrentClient bool) {
	feed := Feedback{
		ErrorCode:    ERROR_CODE_BROADCAST,
		ErrorMessage: "",
		Broadcast:    message.Broadcast,
		Data:         nil,
	}
	feedbyte, _ := feed.Serialization()
	for clientid, client := range hub.Clients {
		if clientid == currentClient.ID && !includeCurrentClient {
			continue
		}
		client.Send <- feedbyte
	}
}

func (hub *TextHub) KickClient(client *Client) {
	close(client.Send)
	delete(hub.Clients, client.ID)
}
