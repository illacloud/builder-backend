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

	// init generator
	echoGenerator := repository.NewEchoGenerator()

	// form echo request by user demand
	componentsList, errInGen := generateBaseComponentPhrase(echoGenerator, userDemand)
	if errInGen != nil {
		fmt.Printf("[ERROR] errInGen: %+v\n", errInGen)
		return nil
	}

	// stack base prompt
	echoGenerator.StackAllHistoryMessage()

	// process single components
	propsFilledComponent := make(map[string]interface{})
	for _, component := range componentsList {
		echoGenerator.CleanHistoryMessages()
		componentAsserted, assertComponentOK := component.(map[string]interface{})
		if !assertComponentOK {
			return errors.New("failed in assert component.")
		}
		displayNameAsserted, assertDisplayNameOK := componentAsserted["displayName"].(string)
		if !assertDisplayNameOK {
			return errors.New("failed in assert displayName.")
		}
		fmt.Printf("[DUMP] now component.type is: %+v\n", componentAsserted["type"])
		generateComponent, errInGenerateComponent := generateComponentPropsPhrase(echoGenerator, componentAsserted)
		if errInGenerateComponent != nil {
			return errInGenerateComponent
		}
		propsFilledComponent[displayNameAsserted] = generateComponent
	}

	fmt.Printf("[DUMP] propsFilledComponent: %+v\n", propsFilledComponent)

	// repack component tree
	componentTree, errInRepack := repackComponentTree(propsFilledComponent)
	if errInRepack != nil {
		return errInRepack
	}
	fmt.Printf("[DUMP] componentTree: %+v\n", componentTree)

	// generate final content
	var finalContent map[string]interface{}
	json.Unmarshal([]byte(componentTree), &finalContent)
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

func generateBaseComponentPhrase(echoGenerator *repository.EchoGenerator, userDemand string) ([]interface{}, error) {
	// generate
	echoGenerator.GenerateBasePrompt(userDemand)
	_, errInEmitEchoRequest := echoGenerator.EmitEchoRequest(false)
	if errInEmitEchoRequest != nil {
		fmt.Printf("[ERROR] errInEmitEchoRequest: %+v\n", errInEmitEchoRequest)
		return nil, errInEmitEchoRequest
	}

	// dump
	historyMessageFinal := echoGenerator.ExportLastHistoryMessages()
	fmt.Printf("\n[DUMP] historyMessageFinal.Content: %+v\n\n", historyMessageFinal.Content)

	// unmarshal
	finalContent, errInUnmarshal := historyMessageFinal.UnMarshalArrayContent()
	if errInUnmarshal != nil {
		return nil, errInUnmarshal
	}
	return finalContent, nil
}

func generateComponentPropsPhrase(echoGenerator *repository.EchoGenerator, component map[string]interface{}) (map[string]interface{}, error) {
	echoGenerator.FillPropsBySingleComponent(component)
	historyMessage1 := echoGenerator.ExportLastHistoryMessages()
	fmt.Printf("\n[DUMP] historyMessage1.Content: %+v\n\n", historyMessage1.Content)
	_, errInEmitEchoRequest := echoGenerator.EmitEchoRequest(false)
	if errInEmitEchoRequest != nil {
		fmt.Printf("[ERROR] errInEmitEchoRequest: %+v\n", errInEmitEchoRequest)
		return nil, errInEmitEchoRequest
	}
	// dump
	historyMessageFinal := echoGenerator.ExportLastHistoryMessages()
	fmt.Printf("\n[DUMP] historyMessageFinal.Content: %+v\n\n", historyMessageFinal.Content)
	unmarshaledContent, errInUnmarshal := historyMessageFinal.UnMarshalObjectContent()
	if errInUnmarshal != nil {
		return nil, errInUnmarshal
	}
	return unmarshaledContent, nil
}

func repackComponentTree(fullComponentList map[string]interface{}) (string, error) {
	// pick up root node
	var rootNode string
	for displayName, component := range fullComponentList {
		componentAsserted, _ := component.(map[string]interface{})
		parentNode, _ := componentAsserted["parentNode"].(string)
		if parentNode == "bodySection1-bodySectionContainer1" {
			rootNode = displayName
		}
	}
	if rootNode == "" {
		return "", errors.New("can not find root node.")
	}

	// recrusive fill
	rootNodeAsserted, _ := fullComponentList[rootNode].(map[string]interface{})
	rootNodePrototype := repository.NewWidgetPrototypeByMap(rootNodeAsserted)
	packComponetRecrusive(rootNodeAsserted, rootNodePrototype, fullComponentList)

	// marshal to json and return it
	componentTreeInJSON, errInMarshal := json.Marshal(rootNodePrototype)
	return string(componentTreeInJSON), errInMarshal
}

func packComponetRecrusive(currentNode map[string]interface{}, currentNodePrototype *repository.WidgetPrototype, fullComponentList map[string]interface{}) {
	childrenNodes, _ := currentNode["childrenNode"].([]interface{})
	if len(childrenNodes) == 0 {
		return
	}
	filledClildrenNodes := make([]interface{}, 0)
	for _, displayName := range childrenNodes {
		displayNameAsserted, _ := displayName.(string)
		filledClildrenNodes = append(filledClildrenNodes, fullComponentList[displayNameAsserted])
	}

	for _, childrenNode := range filledClildrenNodes {
		childrenNodeAsserted, assertChildrenNodeOK := childrenNode.(map[string]interface{})
		if !assertChildrenNodeOK {
			continue
		}
		childrenNodePrototyle := repository.NewWidgetPrototypeByMap(childrenNodeAsserted)
		packComponetRecrusive(childrenNodeAsserted, childrenNodePrototyle, fullComponentList)
		currentNodePrototype.AppendChildrenNode(childrenNodePrototyle)
	}
	return
}
