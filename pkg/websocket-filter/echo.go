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

	"github.com/google/uuid"

	"github.com/illacloud/builder-backend/internal/repository"
	ws "github.com/illacloud/builder-backend/internal/websocket"
)

// - generate base component tree
// - is first time generate
//   - [true]
//   - generate base component
//   - stack base generated component
//   - [false]
//   - re-generate base component
//   - replace base generated component
//
// - generate props
// - have history component
//   - [false]
//   - just generate props by iteration
//   - [true]
//   - merge components props by displayName
//   - generate props
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
	echoGenerator := currentClient.EchoGenerator
	echoGenerator.CleanHistoryMessages() // must do this

	// delete old components
	removeOldComponents(currentClient, hub, echoGenerator)

	// form echo request by user demand
	fmt.Printf("\n- [form echo request by user demand] -------------------------------------------------------------------------\n")
	var componentsList []interface{}
	if !echoGenerator.HaveStackendMessages() {
		fmt.Printf("[DUMP] now is first time generate. run generateBaseComponentPhrase()\n")
		// fisrt time generate
		var errInGen error
		componentsList, errInGen = generateBaseComponentPhrase(echoGenerator, userDemand)
		if errInGen != nil {
			fmt.Printf("[ERROR] errInGen: %+v\n", errInGen)
			return nil
		}
		// stack base prompt
		echoGenerator.StackAllHistoryMessage()
	} else {
		fmt.Printf("[DUMP] NOT first time generate. run generateRawUserDemand(). \n")
		// update user demand
		var errInGen error
		componentsList, errInGen = generateRawUserDemand(echoGenerator, userDemand)
		if errInGen != nil {
			fmt.Printf("[ERROR] errInGen: %+v\n", errInGen)
			return nil
		}
		// stack base prompt
		echoGenerator.StackAllHistoryMessage()
		echoGenerator.DumpStackendMessages()
		echoGenerator.DumpHistoryMessages()
	}

	// process single components
	fmt.Printf("\n- [process single components] -------------------------------------------------------------------------\n")
	propsFilledComponent := make(map[string]interface{})
	for serial, component := range componentsList {
		fmt.Printf("\n- [process single components (%d/%d) ] -------------------------------------------------------------------------\n", serial+1, len(componentsList))
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
		generateComponent, errInGenerateComponent := generateComponentPropsPhrase(echoGenerator, componentAsserted, "")
		if errInGenerateComponent != nil {
			return errInGenerateComponent
		}
		propsFilledComponent[displayNameAsserted] = generateComponent
	}

	fmt.Printf("[DUMP] propsFilledComponent: %+v\n", propsFilledComponent)

	// repack component tree
	fmt.Printf("\n- [repack component tree] -------------------------------------------------------------------------\n")
	componentTree, componentTreeObject, errInRepack := repackComponentTree(propsFilledComponent)
	if errInRepack != nil {
		return errInRepack
	}
	fmt.Printf("[DUMP] componentTree: %+v\n", componentTree)

	// set top tree node displayName
	echoGenerator.SetLastRootNodeDisplayName(componentTreeObject.ExportDisplayName())

	// generate final content
	var finalContent map[string]interface{}
	json.Unmarshal([]byte(componentTree), &finalContent)

	// send
	createComponent(currentClient, hub, finalContent)

	// end
	fmt.Printf("\n- [FINISH] -------------------------------------------------------------------------\n")
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

func generateRawUserDemand(echoGenerator *repository.EchoGenerator, userDemand string) ([]interface{}, error) {
	echoGenerator.FillRawUserDemand(userDemand)
	// generate
	_, errInEmitEchoRequest := echoGenerator.EmitEchoRequest(false)
	if errInEmitEchoRequest != nil {
		fmt.Printf("[ERROR] errInEmitEchoRequest: %+v\n", errInEmitEchoRequest)
		return nil, errInEmitEchoRequest
	}

	// sence new base prompt and result generated, clean Stacked message
	echoGenerator.CleanStackendMessages()

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

func generateComponentPropsPhrase(echoGenerator *repository.EchoGenerator, component map[string]interface{}, userDeamnd string) (map[string]interface{}, error) {
	echoGenerator.FillPropsBySingleComponent(component, userDeamnd)
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

func repackComponentTree(fullComponentList map[string]interface{}) (string, *repository.WidgetPrototype, error) {
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
		return "", nil, errors.New("can not find root node.")
	}

	// recrusive fill
	rootNodeAsserted, _ := fullComponentList[rootNode].(map[string]interface{})
	rootNodePrototype := repository.NewWidgetPrototypeByMap(rootNodeAsserted)
	packComponetRecrusive(rootNodeAsserted, rootNodePrototype, fullComponentList)

	// marshal to json and return it
	componentTreeInJSON, errInMarshal := json.Marshal(rootNodePrototype)
	return string(componentTreeInJSON), rootNodePrototype, errInMarshal
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

func removeOldComponents(currentClient *ws.Client, hub *ws.Hub, echoGenerator *repository.EchoGenerator) {
	fmt.Printf("\n- [removeOldComponents] -------------------------------------------------------------------------\n")

	// get display name
	rootDisplayName := echoGenerator.ExportLastRootNodeDisplayName()
	if len(rootDisplayName) == 0 {
		return
	}

	// pack websocket message
	payloadData := make([]interface{}, 0)
	payloadData = append(payloadData, rootDisplayName)
	broadcastPayloadData := make(map[string]interface{}, 2)
	broadcastPayloadData["displayNames"] = payloadData
	broadcastPayloadData["source"] = "illa-ai"
	broadcastData := &ws.Broadcast{
		Type:    "components/deleteComponentNodeReducer",
		Payload: broadcastPayloadData,
	}

	illaAIUUID, _ := uuid.Parse("000000a1-0000-0000-0000-000000000001")
	messageData := ws.Message{
		Signal:        ws.SIGNAL_DELETE_STATE,
		Target:        1,
		Option:        1,
		Payload:       payloadData,
		Broadcast:     broadcastData,
		ClientID:      illaAIUUID,
		APPID:         currentClient.GetAPPID(),
		NeedBroadcast: true,
	}
	jsonData, _ := json.Marshal(messageData)
	fmt.Printf("[DUMP] ws message: %s\n", jsonData)

	// send it
	fmt.Printf("\n- [call BroadcastToClientItSelf] -------------------------------------------------------------------------\n")
	hub.BroadcastToClientItSelf(&messageData, currentClient)

}

func createComponent(currentClient *ws.Client, hub *ws.Hub, content map[string]interface{}) {
	fmt.Printf("\n- [createComponent] -------------------------------------------------------------------------\n")
	payloadData := make([]interface{}, 0)
	payloadData = append(payloadData, content)
	broadcastData := &ws.Broadcast{
		Type:    "components/addComponentReducer",
		Payload: payloadData,
	}
	illaAIUUID, _ := uuid.Parse("000000a1-0000-0000-0000-000000000001")

	messageData := ws.Message{
		Signal:        ws.SIGNAL_CREATE_STATE,
		Target:        1,
		Option:        1,
		Payload:       payloadData,
		Broadcast:     broadcastData,
		ClientID:      illaAIUUID,
		APPID:         currentClient.GetAPPID(),
		NeedBroadcast: true,
	}
	jsonData, _ := json.Marshal(messageData)
	fmt.Printf("[DUMP] ws message: %s\n", jsonData)

	// send it
	fmt.Printf("\n- [call BroadcastToClientItSelf] -------------------------------------------------------------------------\n")
	hub.BroadcastToClientItSelf(&messageData, currentClient)
}
