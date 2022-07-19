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

import "encoding/json"

type Feedback struct {
	ErrorCode    int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	Broadcast    *Broadcast  `json:"broadcast"`
	Data         interface{} `json:"data"`
}

func (feed *Feedback) Serialization() ([]byte, error) {
	return json.Marshal(feed)
}

func FeedbackCurrentClient(message *Message, currentClient *Client, errorCode int, errorMessage error) {
	feedCurrentClient := Feedback{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage.Error(),
		Broadcast:    message.Broadcast,
		Data:         nil,
	}
	feedbyte, _ := feedCurrentClient.Serialization()
	currentClient.Send <- feedbyte
}

func BroadcastToOtherClients(hub *Hub, message *Message, currentClient *Client) {
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
		if client.RoomID != currentClient.RoomID {
			continue
		}
		client.Send <- feedbyte
	}
}
