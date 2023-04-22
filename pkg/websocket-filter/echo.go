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
	"fmt"

	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/internal/tokenvalidator"
	ws "github.com/illacloud/builder-backend/internal/websocket"
)

func SignalEcho(hub *ws.Hub, message *ws.Message) error {
	currentClient, hit := hub.Clients[message.ClientID]
	if !hit {
		return errors.New("[SignalEcho] target client(" + message.ClientID.String() + ") does dot exists.")
	}
	fmt.Printf("currentClient.ID: %v\n", currentClient.ID)

	// format user demand
	userDemand := ""
	for _, displayNameInterface := range message.Payload {
		userDemand += displayNameInterface.(string)
	}
	if len(userDemand) == 0 {
		return errors.New("[SignalEcho] empty payload")
	}

	// form echo request by user demand
	echoGenerator := repository.NewEchoGenerator()
	userDemandPrompt := echoGenerator.GenerateBasePrompt(userDemand)
	echoMessage := repository.NewEchoMessage(userDemandPrompt)
	echoRequest := repository.NewEchoRequest()
	echoRequest.SetMessage(echoMessage)
	echoPeripheralRequest := repository.NewEchoPeripheralRequest(echoRequest.Export())
	tokenValidator := tokenvalidator.NewRequestTokenValidator()
	token := tokenValidator.GenerateValidateToken(echoPeripheralRequest.Message)
	echoPeripheralRequest.SetValidateToken(token)
	// call echo API
	echoFeedback, _ := repository.Echo(echoPeripheralRequest)
	fmt.Printf("[DUMP] echoFeedback: %+v\n", echoFeedback)
	// check if props not exists, fill props

	// filter components

	// save components

	// broadcast to all clients

	return nil
}
