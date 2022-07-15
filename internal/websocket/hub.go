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
	"errors"
	"fmt"

	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/state"
	"github.com/illa-family/builder-backend/pkg/user"
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

// @todo: the client should check userID, make sure do not broadcast to self.
func (hub *Hub) Run() {
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
			SignalFilter(hub, message)
		}

	}
}

func SignalFilter(hub *Hub, message *Message) error {
	switch message.Signal {
	case SIGNAL_PING:
		return SignalPing(hub, message)
	case SIGNAL_ENTER:
		return SignalEnter(hub, message)
	case SIGNAL_LEAVE:
		return SignalLeave(hub, message)
	case SIGNAL_CREATE_STATE:
		return SignalCreateState(hub, message)
	case SIGNAL_DELETE_STATE:
	case SIGNAL_UPDATE_STATE:
	case SIGNAL_MOVE_STATE:
	case SIGNAL_CREATE_OR_UPDATE:
	case SIGNAL_ONLY_BROADCAST:
	default:
		return nil

	}
	return nil
}

func OptionFilter(hub *Hub, client *Client, message *Message) error {
	return nil
}

func SignalPing(hub *Hub, message *Message) error {
	feed := Feedback{
		ErrorCode:    ERROR_CODE_PONG,
		ErrorMessage: "",
		Broadcast:    nil,
		Data:         nil,
	}
	var feedbyte []byte
	var err error
	if feedbyte, err = feed.Serialization(); err != nil {
		return err
	}
	// send feedback to client itself
	currentClient := hub.Clients[message.ClientID]
	currentClient.Send <- feedbyte
	return nil
}

func SignalEnter(hub *Hub, message *Message) error {
	// init
	currentClient := hub.Clients[message.ClientID]
	var ok bool
	if len(message.Payload) == 0 {
		FeedbackLogInFailed(currentClient)
		return errors.New("[websocket-server] websocket protocol syntax error.")
	}
	var authToken map[string]interface{}
	if authToken, ok = message.Payload[0].(map[string]interface{}); !ok {
		FeedbackLogInFailed(currentClient)
		return errors.New("[websocket-server] websocket protocol syntax error.")
	}
	token, _ := authToken["authToken"].(string)

	// convert authToken to uid
	userID, extractErr := user.ExtractUserIDFromToken(token)
	if extractErr != nil {
		return extractErr
	}
	validAccessToken, validaAccessErr := user.ValidateAccessToken(token)
	if validaAccessErr != nil {
		FeedbackLogInFailed(currentClient)
		return validaAccessErr
	}
	if !validAccessToken {
		FeedbackLogInFailed(currentClient)
		return errors.New("[websocket-server] access token invalied.")
	}
	// assign logged in and mapped user id
	currentClient.IsLoggedIn = true
	currentClient.MappedUserID = userID
	FeedbackLoggedIn(currentClient)
	return nil

}

func SignalLeave(hub *Hub, message *Message) error {
	currentClient := hub.Clients[message.ClientID]
	KickClient(hub, currentClient)
	return nil
}

func SignalCreateState(hub *Hub, message *Message) error {
	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		// build component tree from json
		apprefid := currentClient.RoomID
		summitnodeid := repository.TREE_STATE_SUMMIT_ID
		fmt.Printf("[DUMP] message.Payload: %v \n", message.Payload)
		for _, v := range message.Payload {
			var componenttree *repository.ComponentNode
			componenttree = repository.ConstructComponentNodeByMap(v)
			fmt.Printf("%v\n", componenttree)
			if err := hub.TreeStateServiceImpl.CreateComponentTree(apprefid, summitnodeid, componenttree); err != nil {
				return err
			}
		}
		// feedback
		return nil
	case TARGET_DEPENDENCIES:
	case TARGET_DRAG_SHADOW:
	case TARGET_DOTTED_LINE_SQUARE:
	case TARGET_DISPLAY_NAME:
	case TARGET_APPS:
	case TARGET_RESOURCE:
	}
	// feedback current client
	// broadcast to all room client
	return nil
}

func SignalDeleteState(hub *Hub, message *Message) error {
	return nil
}
func SignalUpdateState(hub *Hub, message *Message) error {
	return nil
}
func SignalMoveState(hub *Hub, message *Message) error {
	return nil
}
func SignalCreateOrUpdateupdate(hub *Hub, message *Message) error {
	return nil
}
func SignalOnlyBroadcast(hub *Hub, message *Message) error {
	return nil
}

func KickClient(hub *Hub, client *Client) {
	close(client.Send)
	delete(hub.Clients, client.ID)
}

func FeedbackLoggedIn(client *Client) {
	// send feedback
	feed := Feedback{
		ErrorCode:    ERROR_CODE_LOGGEDIN,
		ErrorMessage: "",
		Broadcast:    nil,
		Data:         nil,
	}
	var feedbyte []byte
	var err error
	if feedbyte, err = feed.Serialization(); err != nil {
		return
	}
	client.Send <- feedbyte
}

func FeedbackLogInFailed(client *Client) {
	// send feedback
	feed := Feedback{
		ErrorCode:    ERROR_CODE_LOGIN_FAILED,
		ErrorMessage: "",
		Broadcast:    nil,
		Data:         nil,
	}
	var feedbyte []byte
	var err error
	if feedbyte, err = feed.Serialization(); err != nil {
		return
	}
	client.Send <- feedbyte
}
