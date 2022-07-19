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
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/resource"
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
			filter.SignalFilter(hub, message)
		}

	}
}

func SignalFilter(hub *Hub, message *Message) error {
	switch message.Signal {
	case SIGNAL_PING:
		return filter.SignalPing(hub, message)
	case SIGNAL_ENTER:
		return filter.SignalEnter(hub, message)
	case SIGNAL_LEAVE:
		return filter.SignalLeave(hub, message)
	case SIGNAL_CREATE_STATE:
		return filter.SignalCreateState(hub, message)
	case SIGNAL_DELETE_STATE:
		return filter.SignalDeleteState(hub, message)
	case SIGNAL_UPDATE_STATE:
		return filter.SignalUpdateState(hub, message)
	case SIGNAL_MOVE_STATE:
		return filter.SignalMoveState(hub, message)
	case SIGNAL_CREATE_OR_UPDATE:
		return filter.SignalCreateOrUpdate(hub, message)
	case SIGNAL_ONLY_BROADCAST:
		return filter.SignalOnlyBroadcast(hub, message)
	default:
		return nil

	}
	return nil
}

func OptionFilter(hub *Hub, client *Client, message *Message) error {
	return nil
}



func SignalEnter(hub *Hub, message *Message) error {
	// init
	currentClient := hub.Clients[message.ClientID]
	var ok bool
	if len(message.Payload) == 0 {
		errorMessage := errors.New("[websocket-server] websocket protocol syntax error.")
		FeedbackLogInFailed(currentClient)
		FeedbackCurrentClient(message, currentClient, ERROR_CODE_LOGIN_FAILED, errorMessage)
		return errorMessage
	}
	var authToken map[string]interface{}
	if authToken, ok = message.Payload[0].(map[string]interface{}); !ok {
		errorMessage := errors.New("[websocket-server] websocket protocol syntax error.")
		FeedbackCurrentClient(message, currentClient, ERROR_CODE_LOGIN_FAILED, errorMessage)
		return errorMessage
	}
	token, _ := authToken["authToken"].(string)

	// convert authToken to uid
	userID, extractErr := user.ExtractUserIDFromToken(token)
	if extractErr != nil {
		return extractErr
	}
	validAccessToken, validaAccessErr := user.ValidateAccessToken(token)
	if validaAccessErr != nil {
		FeedbackCurrentClient(message, currentClient, ERROR_CODE_LOGIN_FAILED, validaAccessErr)
		return validaAccessErr
	}
	if !validAccessToken {
		errorMessage := errors.New("[websocket-server] access token invalied.")
		FeedbackCurrentClient(message, currentClient, ERROR_CODE_LOGIN_FAILED, errorMessage)
		return errorMessage
	}
	// assign logged in and mapped user id
	currentClient.IsLoggedIn = true
	currentClient.MappedUserID = userID
	FeedbackCurrentClient(message, currentClient, ERROR_CODE_LOGGEDIN)
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
	stateType := repository.STATE_TYPE_INVALIED
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		// build component tree from json
		
		summitnodeid := repository.TREE_STATE_SUMMIT_ID
		
		for _, v := range message.Payload {
			var componenttree *repository.ComponentNode
			componenttree = repository.ConstructComponentNodeByMap(v)
			
			if err := hub.TreeStateServiceImpl.CreateComponentTree(apprefid, summitnodeid, componenttree); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
				return err
			}
		}
	case TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// create k-v state
		
		
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType

			if _, err := hub.KVStateServiceImpl.CreateKVState(kvstatedto); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
				return err
			}
		}
	case TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		// create set state
		
		
		for _, v := range message.Payload {
			// resolve payload
			dns, err := repository.ConstructDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// save state
			for _, displayName := range dns {
				var setStateDto state.SetStateDto
				setStateDto.ConstructByValue(displayName)
				setStateDto.StateType = stateType
				// create state
				if _, err := hub.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
			}
		}
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}
	// feedback currentClient
	FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_OK)

	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)
	return nil
}

func SignalDeleteState(hub *Hub, message *Message) error {
	

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)

	message.RewriteBroadcast()

	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		for _, v := range message.Payload {
			var nowNode state.TreeStateDto
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS
			

			if err := hub.TreeStateServiceImpl.DeleteTreeStateNodeRecursive(apprefid, &nowNode); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_DELETE_STATE_FAILED)
				return err
			}
		}

	case TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// delete k-v state
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType
			

			if err := hub.KVStateServiceImpl.DeleteKVStateByKey(apprefid, &kvstatedto); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_DELETE_STATE_FAILED)
				return err
			}
		}

	case TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		// create dnsplayName state
		
		
		for _, v := range message.Payload {
			// resolve payload
			dns, err := repository.ConstructDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// save state
			for _, displayName := range dns {
				// init
				var setStateDto state.SetStateDto
				setStateDto.ConstructByValue(displayName)
				setStateDto.StateType = stateType
				setStateDto.AppRefID = apprefid
				setStateDto.Version = repository.APP_EDIT_VERSION
				// delete state
				if err := hub.SetStateServiceImpl.DeleteSetStateByValue(&setStateDto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
			}
		}
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	FeedbackCurrentClient(message, currentClient, ERROR_DELETE_STATE_OK)
	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)
	return nil
}

func SignalUpdateState(hub *Hub, message *Message) error {
	

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		for _, v := range message.Payload {
			// update
			// construct update data
			var nowNode state.TreeStateDto
			componentNode := repository.ConstructComponentNodeByMap(v)
			

			serializedComponent, err := componentNode.SerializationForDatabase()
			if err != nil {
				return err
			}
			
			nowNode.Content = string(serializedComponent)
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS
			
			// update
			if err := hub.TreeStateServiceImpl.UpdateTreeStateNode(apprefid, &nowNode); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
				return err
			}
		}

		// feedback currentClient
		FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_OK)

		// feedback otherClient
		BroadcastToOtherClients(hub, message, currentClient)
	case TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// update K-V State
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType
			
			// update
			if err := hub.KVStateServiceImpl.UpdateKVStateByKey(apprefid, &kvstatedto); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
				return err
			}
		}
	case TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			// resolve payload
			dnsfu, err := repository.ConstructDisplayNameStateForUpdateByPayload(v)
			if err != nil {
				return err
			}
			// init state dto
			var beforeSetStateDto state.SetStateDto
			var afterSetStateDto state.SetStateDto
			beforeSetStateDto.ConstructByDisplayNameForUpdate(dnsfu)
			beforeSetStateDto.StateType = stateType
			beforeSetStateDto.AppRefID = apprefid
			beforeSetStateDto.Version = repository.APP_EDIT_VERSION
			afterSetStateDto.ConstructByDisplayNameForUpdate(dnsfu)
			// update state
			if err := hub.SetStateServiceImpl.UpdateSetStateByValue(beforeSetStateDto, afterSetStateDto); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
				return err
			}
		}
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_OK)

	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)

	return nil
}

func SignalMoveState(hub *Hub, message *Message) error {
	

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		apprefid := currentClient.RoomID
		for _, v := range message.Payload {
			var nowNode state.TreeStateDto
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS
			
			if err := hub.TreeStateServiceImpl.MoveTreeStateNode(apprefid, &nowNode); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_MOVE_STATE_FAILED)
				return err
			}
		}

		// feedback currentClient
		FeedbackCurrentClient(message, currentClient, ERROR_MOVE_STATE_OK)

		// feedback otherClient
		BroadcastToOtherClients(hub, message, currentClient)

	case TARGET_DEPENDENCIES:
		fallthrough
	case TARGET_DRAG_SHADOW:
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		FeedbackCurrentClient(message, currentClient, ERROR_CAN_NOT_MOVE_KVSTATE)
		return nil
	case TARGET_DISPLAY_NAME:
		FeedbackCurrentClient(message, currentClient, ERROR_CAN_NOT_MOVE_SETSTATE)
		return nil
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}
	return nil
}

func SignalCreateOrUpdate(hub *Hub, message *Message) error {
	

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	apprefid := 
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		for _, v := range message.Payload {
			// check if state already in database
			var nowNode state.TreeStateDto
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS
			isStateExists := hub.TreeStateServiceImpl.IsTreeStateNodeExists(apprefid, &nowNode)
			if !isStateExists {
				// create
				summitnodeid := repository.TREE_STATE_SUMMIT_ID
				var componenttree *repository.ComponentNode
				componenttree = repository.ConstructComponentNodeByMap(v)
				
				if err := hub.TreeStateServiceImpl.CreateComponentTree(apprefid, summitnodeid, componenttree); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
			} else {
				// update
				// construct update data
				var nowNode state.TreeStateDto
				componentNode := repository.ConstructComponentNodeByMap(v)
				

				serializedComponent, err := componentNode.SerializationForDatabase()
				if err != nil {
					return err
				}
				
				nowNode.Content = string(serializedComponent)
				nowNode.ConstructByMap(v) // set Name
				nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS
				
				// update
				if err := hub.TreeStateServiceImpl.UpdateTreeStateNode(apprefid, &nowNode); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
					return err
				}
			}

		}
	case TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType
			

			isStateExists := hub.KVStateServiceImpl.IsKVStateNodeExists(apprefid, &kvstatedto)
			if !isStateExists {
				if _, err := hub.KVStateServiceImpl.CreateKVState(kvstatedto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
			} else {
				// update
				if err := hub.KVStateServiceImpl.UpdateKVStateByKey(apprefid, &kvstatedto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
					return err
				}
			}

		}
	case TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			// resolve payload
			dns, err := repository.ConstructDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// create or update state
			for _, displayName := range dns {
				// checkout
				var setStateDto state.SetStateDto
				var setStateDtoInDB *state.SetStateDto
				var err error
				setStateDto.ConstructByValue(displayName)
				setStateDto.ConstructByType(stateType)
				setStateDto.ConstructByApp(appDto)
				setStateDto.ConstructWithEditVersion()
				// lookup state
				if setStateDtoInDB, err = hub.SetStateServiceImpl.GetByValue(setStateDto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
				if setStateDtoInDB == nil {
					// create
					if _, err = hu.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
						FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
						return err
					}
				} else {
					// update
					setStateDtoInDB.ConstructByValue(setStateDto.Value)
					if _, err = hu.SetStateServiceImpl.UpdateSetState(setStateDtoInDB); err != nil {
						FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
						return err
					}
				}
			}
		}
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	FeedbackCurrentClient(message, currentClient, ERROR_CREATE_OR_UPDATE_STATE_OK)

	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)
	return nil
}

func SignalOnlyBroadcast(hub *Hub, message *Message) error {
	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	message.RewriteBroadcast()

	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)
	return nil
}

func KickClient(hub *Hub, client *Client) {
	close(client.Send)
	delete(hub.Clients, client.ID)
}
