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

// clients hub, maintains active clients and broadcast messags.
type Hub struct {
	// registered clients map
	Clients map[*Client]bool

	// inbound messages from the clients.
	// try ```hub.Broadcast <- []byte(message)```
	Broadcast chan []byte

	// on message process
	OnMessage chan *Message

	// register requests from the clients.
	Register chan *Client

	// unregister requests from the clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		OnMessage:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// @todo: the client should check userID, make sure do not broadcast to self.
func (hub *Hub) Run() {
	for {
		select {
		// handle register event
		case client := <-hub.Register:
			hub.Clients[client] = true
		// handle unregister events
		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.Send)
			}
		// handle all hub broadcast events
		case message := <-hub.Broadcast:
			for client := range hub.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(hub.Clients, client)
				}
			}
		// handle client on message event
		case message := <-hub.OnMessage:
			for client := range hub.Clients {
				// check room ID
				if client.RoomID != message.RoomID {
					continue
				}
				if !SignalFilter(hub, client, message) {
					continue
				}
				// hub option filter
				if !OptionFilter(hub, client, message) {
					continue
				}
			}
		}

	}
}

func SignalFilter(hub *Hub, client *Client, message *Message) bool {
	switch message.Protocol.Signal {
	case SIGNAL_LOGIN:
		return true
	case SIGNAL_LOGOUT:
		close(client.Send)
		delete(hub.Clients, client)
		return false
	default:
		return false

	}

}

func OptionFilter(hub *Hub, client *Client, message *Message) bool {
	if message.Protocol.Option&OPTION_BROADCAST_ROOM == message.Protocol.Option {
		// room boardcast
		select {
		case client.Send <- message.RawProtocol:
		default:
			close(client.Send)
			delete(hub.Clients, client)
		}
		return true
	}
	return false
}
