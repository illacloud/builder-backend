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
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/websocket"
	"google.golang.org/protobuf/proto"
)

func SignalMoveStateBinary(hub *websocket.Hub, message *websocket.MovingMessageBin) error {

	// deserialize message
	clientIDString := message.GetClientID()
	clientID, errInParseClientID := uuid.Parse(clientIDString)
	if errInParseClientID != nil {
		return errInParseClientID
	}
	currentClient, hit := hub.BinaryClients[clientID]
	if !hit {
		return errors.New("[SignalMoveStateBinary] target client(" + message.ClientID + ") does dot exists.")
	}
	// feedback otherClient
	binaryMessage, errInMarshal := proto.Marshal(message)
	if errInMarshal != nil {
		log.Printf("[SignalMoveStateBinary] Failed to encode MovingMessageBin: ", errInMarshal)
		return errInMarshal
	}
	hub.BroadcastBinaryToOtherClients(binaryMessage, currentClient)

	return nil
}
