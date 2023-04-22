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
	"encoding/json"
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
	echoGenerator.GenerateBasePrompt(userDemand)
	fullHistoryMessage := echoGenerator.ExportFullHistoryMessages()
	echoRequest := repository.NewEchoRequest()
	echoRequest.SetMessages(fullHistoryMessage)
	echoPeripheralRequest := repository.NewEchoPeripheralRequest(echoRequest.Export())
	tokenValidator := tokenvalidator.NewRequestTokenValidator()
	token := tokenValidator.GenerateValidateToken(echoPeripheralRequest.Message)
	echoPeripheralRequest.SetValidateToken(token)

	// call echo API
	echoFeedback, _ := repository.Echo(echoPeripheralRequest)
	fmt.Printf("[DUMP] echoFeedback: %+v\n", echoFeedback)
	historyMessage, queryDidNotFinish, errInExportMessage := echoFeedback.ExportMessage()
	if errInExportMessage != nil {
		fmt.Printf("[ERROR] errInExportMessage: %+v\n", errInExportMessage)
		return errInExportMessage
	}
	// @todo: process if query did not finish
	fmt.Printf("[queryDidNotFinish?] %v\n", queryDidNotFinish)

	// log history message
	echoGenerator.SaveHistoryMessage(historyMessage)

	// check if props not exists, fill props
	componentTypeList := historyMessage.DetectComponentTypes()
	fmt.Printf("[DUMP] componentTypeList: %+v\n", componentTypeList)
	echoGenerator.FillPropsByContext(componentTypeList)

	// generate request again
	echoRequest2 := repository.NewEchoRequest()
	fullHistoryMessage2 := echoGenerator.ExportFullHistoryMessages()
	echoRequest2.SetMessages(fullHistoryMessage2)
	echoPeripheralRequest2 := repository.NewEchoPeripheralRequest(echoRequest2.Export())
	token2 := tokenValidator.GenerateValidateToken(echoPeripheralRequest2.Message)
	echoPeripheralRequest2.SetValidateToken(token2)

	// call API again
	echoFeedback2, _ := repository.Echo(echoPeripheralRequest2)
	fmt.Printf("[DUMP] echoFeedback: %+v\n", echoFeedback2)
	historyMessageFinal, _, _ := echoFeedback.ExportMessage()
	// filter components

	// save components

	// new message
	// finalContent, _ := historyMessage.UnMarshalContent()
	finalContent, _ := historyMessageFinal.UnMarshalContent()
	payloadData := make([]interface{}, 0)
	payloadData = append(payloadData, finalContent)
	broadcastData := &ws.Broadcast{
		Type:    "components/addComponentReducer/remote",
		Payload: payloadData,
	}

	messageData := ws.Message{
		ClientID:      currentClient.GetID(),
		Signal:        ws.SIGNAL_CREATE_STATE,
		APPID:         currentClient.GetAPPID(),
		Option:        1,
		Payload:       payloadData,
		Target:        1,
		Broadcast:     broadcastData,
		NeedBroadcast: true,
	}
	jsonData, _ := json.Marshal(messageData)
	fmt.Printf("[DUMP] ws message: %s\n", jsonData)

	// broadcast to all clients
	hub.BroadcastToRoomAllClients(&messageData, currentClient)

	return nil
}
