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

package websocket

import (
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/resource"
	"github.com/illa-family/builder-backend/pkg/state"
	uuid "github.com/satori/go.uuid"
)

// clients hub, maintains active clients and broadcast messags.
type Hub struct {
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

	// impl
	TreeStateServiceImpl *state.TreeStateServiceImpl
	KVStateServiceImpl   *state.KVStateServiceImpl
	SetStateServiceImpl  *state.SetStateServiceImpl
	AppServiceImpl       *app.AppServiceImpl
	ResourceServiceImpl  *resource.ResourceServiceImpl
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID]*Client),
		Broadcast:  make(chan []byte),
		OnMessage:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
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

func (hub *Hub) BroadcastToOtherClients(message *Message, currentClient *Client) {
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
