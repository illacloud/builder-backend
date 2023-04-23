package repository

import (
	"encoding/json"
	"fmt"

	"github.com/illacloud/builder-backend/internal/tokenvalidator"
)

// role

const (
	ROLE_USER      = "user"
	ROLE_ASSISTANT = "assistant"
)

// template base prompt

const (
	TEMPLATE_BASE_PROMPT_COMPONENT_SCHEMA         = "consider a json struct named component like {type:\"\", \"containerType\": \"EDITOR_SCALE_SQUARE\", displayName:\"\",parentNode:\"\",childrenNode:[],h:0,w:0,x:0,y:0,props:{}}. "
	TEMPLATE_BASE_PROMPT_COMPONENT_TYPE           = "component type are in CONTAINER_WIDGET, FORM_WIDGET, MODAL_WIDGET, CANVAS, TABLE_WIDGET, TEXT_WIDGET, BUTTON_WIDGET, INPUT_WIDGET, NUMBER_INPUT_WIDGET, SELECT_WIDGET, CHART_WIDGET, IMAGE_WIDGET, UPLOAD_WIDGET, EDITABLE_TEXT_WIDGET, SLIDER_WIDGET, RANGE_SLIDER_WIDGET, SWITCH_WIDGET, MULTISELECT_WIDGET, CHECKBOX_GROUP_WIDGET. the CONTAINER_WIDGET, FORM_WIDGET, MODAL_WIDGET must have only one childrenNode CANVAS, and CANVAS childrenNode can contain other widget. "
	TEMPLATE_BASE_PROMPT_COMPONENT_CONTAINER_TYPE = "containerType is fixed EDITOR_SCALE_SQUARE . "
	TEMPLATE_BASE_PROMPT_COMPONENT_DISPLAYNAME    = "displayName value is type field concat serial number with \"_\" and global unique. "
	TEMPLATE_BASE_PROMPT_COMPONENT_PARENT_NODE    = "top level parentNode value must be \"bodySection1-bodySectionContainer1\". "
	TEMPLATE_BASE_PROMPT_COMPONENT_HWXY           = "all components are rectangle. h, w are component size. w should not above 60. x, y are left-top position of component, should start with 0. "
	TEMPLATE_BASE_PROMPT_COMPONENT_PROPS          = "props leave it as an empty json object. "
	TEMPLATE_BASE_PROMPT_COMPONENT_GENERATE       = "%s, no prose, no note, output only JSON object. "
)

// components base prompt
const (
	COMPONENTS_BASE_PROMPT                       = "now fill component props field with reasonable data. "
	COMPONENTS_BASE_PROMPT_CONTAINER_WIDGET      = "{\"$dynamicAttrPaths\": [],\"backgroundColor\": \"#f0f9ffff\",\"borderColor\": \"#ffffffff\",\"borderWidth\": \"0px\",\"currentIndex\": 0,\"currentKey\": \"View 1\",\"dynamicHeight\": \"fixed\",\"radius\": \"4px\",\"resizeDirection\": \"ALL\",\"shadow\": \"small\"}"
	COMPONENTS_BASE_PROMPT_FORM_WIDGET           = "{\"showHeader\": true,\"showFooter\": true,\"validateInputsOnSubmit\": true,\"resetAfterSuccessful\": true,\"borderColor\": \"#ffffffff\",\"backgroundColor\": \"#ffffffff\",\"radius\": \"4px\",\"borderWidth\": \"4px\",\"shadow\": \"small\",\"headerHeight\": 11,\"footerHeight\": 7,\"$dynamicAttrPaths\": []}"
	COMPONENTS_BASE_PROMPT_MODAL_WIDGET          = "{\"backgroundColor\": \"#ffffffff\",\"borderColor\": \"#ffffffff\",\"borderWidth\": \"1px\",\"clickMaskClose\": false,\"footerHeight\": 7,\"headerHeight\": 11,\"radius\": \"4px\",\"shadow\": \"small\",\"showFooter\": true,\"showHeader\": true}"
	COMPONENTS_BASE_PROMPT_CANVAS                = "{\"viewList\":[{\"id\":\"1488fe5a-85f2-4a93-a9c8-1efcae9ea51c\",\"key\":\"View1\",\"label\":\"View1\"},{\"id\":\"d128dd71-6252-41c1-b27d-6a85b90267b1\",\"key\":\"View2\",\"label\":\"View2\"},{\"id\":\"6ac0b744-ad8b-4504-ad32-c64aa68bb8db\",\"key\":\"View3\",\"label\":\"View3\"}],\"currentIndex\":2,\"currentKey\":\"View3\",\"borderColor\":\"#ffffffff\",\"backgroundColor\":\"#ffffffff\",\"radius\":\"4px\",\"borderWidth\":\"1px\",\"shadow\":\"small\",\"dynamicHeight\":\"fixed\",\"resizeDirection\":\"ALL\",\"$dynamicAttrPaths\":[]}}"
	COMPONENTS_BASE_PROMPT_TABLE_WIDGET          = "{\"$dynamicAttrPaths\":[],\"columns\":[{\"accessorKey\":\"id\",\"columnIndex\":0,\"enableSorting\":true,\"header\":\"id\",\"id\":\"id\",\"type\":\"text\",\"visible\":true},{\"accessorKey\":\"name\",\"columnIndex\":0,\"enableSorting\":true,\"header\":\"name\",\"id\":\"name\",\"type\":\"text\",\"visible\":true},{\"accessorKey\":\"email\",\"columnIndex\":0,\"enableSorting\":true,\"header\":\"email\",\"id\":\"email\",\"type\":\"text\",\"visible\":true}],\"dataSourceJS\":\"{{list_all.data}}\",\"dataSourceMode\":\"dynamic\",\"defaultSortKey\":\"id\",\"defaultSortOrder\":\"ascend\",\"download\":false,\"emptyState\":\"Norowsfound\",\"filter\":false,\"overFlow\":\"pagination\",\"pageSize\":\"{{10}}\"}"
	COMPONENTS_BASE_PROMPT_TEXT_WIDGET           = "{\"$dynamicAttrPaths\": [],\"colorScheme\": \"grayBlue\",\"disableMarkdown\": false,\"dynamicHeight\": \"auto\",\"fs\": \"14px\",\"hidden\": false,\"horizontalAlign\": \"start\",\"resizeDirection\": \"HORIZONTAL\",\"value\": \"# Dashboard\",\"verticalAlign\": \"center\"}"
	COMPONENTS_BASE_PROMPT_BUTTON_WIDGET         = "{\"text\": \"Button\",\"variant\": \"fill\",\"colorScheme\": \"blue\",\"hidden\": false,\"$dynamicAttrPaths\": []}"
	COMPONENTS_BASE_PROMPT_INPUT_WIDGET          = "{\"value\": \"\",\"label\": \"Label\",\"labelAlign\": \"left\",\"labelPosition\": \"left\",\"labelWidth\": \"{{33}}\",\"colorScheme\": \"blue\",\"hidden\": false,\"formDataKey\": \"{{input1.displayName}}\",\"placeholder\": \"input sth\",\"$dynamicAttrPaths\": []}"
	COMPONENTS_BASE_PROMPT_NUMBER_INPUT_WIDGET   = "{\"label\": \"Label\",\"labelAlign\": \"left\",\"labelPosition\": \"left\",\"labelWidth\": \"{{33}}\",\"colorScheme\": \"blue\",\"hidden\": false,\"formDataKey\": \"{{numberInput1.displayName}}\",\"$dynamicAttrPaths\": [    \"labelWidth\",    \"formDataKey\"]}"
	COMPONENTS_BASE_PROMPT_SELECT_WIDGET         = "{\"optionConfigureMode\":\"static\",\"label\":\"Label\",\"labelAlign\":\"left\",\"labelPosition\":\"left\",\"labelWidth\":\"{{33}}\",\"manualOptions\":[{\"id\":\"option-db33ac88-6319-4ee0-b922-63dc53b77671\",\"label\":\"Option1\",\"value\":\"Option1\"},{\"id\":\"option-765ca2d5-073b-4677-8a13-327bad08f304\",\"label\":\"Option2\",\"value\":\"Option2\"},{\"id\":\"option-db200246-0423-4540-b972-6b2d9b8d4a56\",\"label\":\"Option3\",\"value\":\"Option3\"}],\"dataSources\":\"{{[]}}\",\"colorScheme\":\"blue\",\"hidden\":false,\"formDataKey\":\"{{select1.displayName}}\",\"$dynamicAttrPaths\":[\"labelWidth\",\"dataSources\",\"formDataKey\"]}"
	COMPONENTS_BASE_PROMPT_CHART_WIDGET          = "{\"dataSourceJS\":\"{{list_all.data}}\",\"chartType\":\"bar\",\"dataSourceMode\":\"dynamic\",\"xAxis\":\"month\",\"datasets\":[{\"id\":\"8e6fc947-f354-4e33-977d-7dd0ca85b23a\",\"datasetName\":\"Dataset1\",\"datasetValues\":\"users\",\"aggregationMethod\":\"SUM\",\"type\":\"bar\",\"color\":\"#165DFF\"}],\"$dynamicAttrPaths\":[]}"
	COMPONENTS_BASE_PROMPT_IMAGE_WIDGET          = "{\"imageSrc\":\"https://images.unsplash.com/photo-1614853316476-de00d14cb1fc?ixlib=rb-1.2.1&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=2370&q=80\",\"radius\":\"0px\",\"hidden\":false,\"objectFit\":\"cover\",\"$dynamicAttrPaths\":[]}"
	COMPONENTS_BASE_PROMPT_UPLOAD_WIDGET         = "{\"type\":\"button\",\"buttonText\":\"Upload\",\"selectionType\":\"single\",\"dropText\":\"Selectordropafilehere\",\"verticalAlign\":\"center\",\"hidden\":false,\"appendFiles\":false,\"fileType\":\"\",\"variant\":\"fill\",\"colorScheme\":\"blue\",\"formDataKey\":\"{{upload1.displayName}}\",\"showFileList\":false,\"sizeType\":\"mb\",\"dynamicHeight\":\"auto\",\"$dynamicAttrPaths\":[]}"
	COMPONENTS_BASE_PROMPT_EDITABLE_TEXT_WIDGET  = "{\"label\": \"Label\",\"labelAlign\": \"left\",\"labelPosition\": \"left\",\"labelWidth\": \"{{33}}\",\"colorScheme\": \"blue\",\"hidden\": false,\"value\": \"editable text for display\",\"$dynamicAttrPaths\": []}"
	COMPONENTS_BASE_PROMPT_SLIDER_WIDGET         = "{\"value\":\"{{2}}\",\"min\":\"{{0}}\",\"max\":\"{{10}}\",\"step\":\"{{1}}\",\"label\":\"Label\",\"labelAlign\":\"left\",\"labelPosition\":\"left\",\"labelWidth\":\"{{33}}\",\"hideOutput\":false,\"disabled\":false,\"colorScheme\":\"blue\",\"hidden\":false,\"formDataKey\":\"{{slider1.displayName}}\",\"$dynamicAttrPaths\":[]}"
	COMPONENTS_BASE_PROMPT_RANGE_SLIDER_WIDGET   = "{\"startValue\": \"{{3}}\",\"endValue\": \"{{7}}\",\"min\": \"{{0}}\",\"max\": \"{{10}}\",\"step\": \"{{1}}\",\"label\": \"Label\",\"labelAlign\": \"left\",\"labelPosition\": \"left\",\"labelWidth\": \"{{33}}\",\"hideOutput\": false,\"disabled\": false,\"colorScheme\": \"blue\",\"hidden\": false,\"formDataKey\": \"{{rangeSlider1.displayName}}\",\"$dynamicAttrPaths\": []}"
	COMPONENTS_BASE_PROMPT_SWITCH_WIDGET         = "{\"label\": \"Label\",\"labelAlign\": \"left\",\"labelPosition\": \"left\",\"labelWidth\": \"{{33}}\",\"labelFull\": \"{{true}}\",\"colorScheme\": \"blue\",\"hidden\": \"{{false}}\",\"$dynamicAttrPaths\": []}"
	COMPONENTS_BASE_PROMPT_MULTISELECT_WIDGET    = "{\"label\":\"Label\",\"optionConfigureMode\":\"static\",\"labelAlign\":\"left\",\"labelPosition\":\"left\",\"labelWidth\":\"{{33}}\",\"dataSources\":\"{{[]}}\",\"colorScheme\":\"blue\",\"hidden\":false,\"manualOptions\":[{\"id\":\"option-73733667-a63f-44ef-9caf-4700d1138cea\",\"label\":\"Option1\",\"value\":\"Option1\"},{\"id\":\"option-3633908a-40b5-4bd3-9530-5fd87a0a760c\",\"label\":\"Option2\",\"value\":\"Option2\"},{\"id\":\"option-1c7c6a83-1a4b-4a42-917c-cb0ff1541ae1\",\"label\":\"Option3\",\"value\":\"Option3\"}],\"dynamicHeight\":\"auto\",\"formDataKey\":\"{{multiselect1.displayName}}\",\"resizeDirection\":\"HORIZONTAL\",\"$dynamicAttrPaths\":[]}"
	COMPONENTS_BASE_PROMPT_CHECKBOX_GROUP_WIDGET = "{\"optionConfigureMode\":\"static\",\"label\":\"Label\",\"labelAlign\":\"left\",\"labelPosition\":\"left\",\"labelWidth\":\"{{33}}\",\"manualOptions\":[{\"id\":\"option-6cd4af1c-16fb-49c8-9098-2abedbb8678f\",\"label\":\"Option1\",\"value\":\"Option1\"},{\"id\":\"option-7cdb88c3-e213-426f-adaf-3c4118b347de\",\"label\":\"Option2\",\"value\":\"Option2\"},{\"id\":\"option-bc940e14-2df5-4cff-84d7-cbeb87b19e8b\",\"label\":\"Option3\",\"value\":\"Option3\"}],\"dataSources\":\"{{[]}}\",\"direction\":\"horizontal\",\"colorScheme\":\"blue\",\"formDataKey\":\"{{checkboxGroup1.displayName}}\",\"$dynamicAttrPaths\":[]}"
)

const (
	PRIMITIVE_PROMPT_CONTINUE = "continue"
)

type EchoGenerator struct {
	Placeholder     string
	HistoryMessages []*HistoryMessage
}

func NewEchoGenerator() *EchoGenerator {
	return &EchoGenerator{}
}

func (egen *EchoGenerator) SaveHistoryMessageInRaw(role string, content string) {
	historyMessage := &HistoryMessage{
		Role:    role,
		Content: content,
	}
	egen.HistoryMessages = append(egen.HistoryMessages, historyMessage)
}

func (egen *EchoGenerator) SaveHistoryMessage(historyMessage *HistoryMessage) {
	egen.HistoryMessages = append(egen.HistoryMessages, historyMessage)
}

func (egen *EchoGenerator) ExportFullHistoryMessages() []*HistoryMessage {
	return egen.HistoryMessages
}

func (egen *EchoGenerator) ExportLastHistoryMessages() *HistoryMessage {
	lastMessageSerial := len(egen.HistoryMessages) - 1
	return egen.HistoryMessages[lastMessageSerial]
}

func (egen *EchoGenerator) GenerateBasePrompt(userDemand string) string {
	ret := fmt.Sprintf(
		TEMPLATE_BASE_PROMPT_COMPONENT_SCHEMA+
			TEMPLATE_BASE_PROMPT_COMPONENT_TYPE+
			TEMPLATE_BASE_PROMPT_COMPONENT_CONTAINER_TYPE+
			TEMPLATE_BASE_PROMPT_COMPONENT_DISPLAYNAME+
			TEMPLATE_BASE_PROMPT_COMPONENT_PARENT_NODE+
			TEMPLATE_BASE_PROMPT_COMPONENT_HWXY+
			TEMPLATE_BASE_PROMPT_COMPONENT_PROPS+
			TEMPLATE_BASE_PROMPT_COMPONENT_GENERATE, userDemand,
	)
	// auto save history message
	egen.SaveHistoryMessageInRaw(ROLE_USER, ret)
	return ret
}

func (egen *EchoGenerator) FillPropsByContext(componentTypeList map[string]bool) string {
	ret := COMPONENTS_BASE_PROMPT
	for componentType, _ := range componentTypeList {
		switch componentType {
		case "CONTAINER_WIDGET":
			ret += "CONTAINER_WIDGET props be like " + COMPONENTS_BASE_PROMPT_CONTAINER_WIDGET + ". "
		case "FORM_WIDGET":
			ret += "FORM_WIDGET props be like " + COMPONENTS_BASE_PROMPT_FORM_WIDGET + ". "
		case "MODAL_WIDGET":
			ret += "MODAL_WIDGET props be like " + COMPONENTS_BASE_PROMPT_MODAL_WIDGET + ". "
		case "CANVAS":
			ret += "CANVAS props be like " + COMPONENTS_BASE_PROMPT_CANVAS + ". "
		case "TABLE_WIDGET":
			ret += "TABLE_WIDGET props be like " + COMPONENTS_BASE_PROMPT_TABLE_WIDGET + ". "
		case "TEXT_WIDGET":
			ret += "TEXT_WIDGET props be like " + COMPONENTS_BASE_PROMPT_TEXT_WIDGET + ". "
		case "BUTTON_WIDGET":
			ret += "BUTTON_WIDGET props be like " + COMPONENTS_BASE_PROMPT_BUTTON_WIDGET + ". "
		case "INPUT_WIDGET":
			ret += "INPUT_WIDGET props be like " + COMPONENTS_BASE_PROMPT_INPUT_WIDGET + ". "
		case "NUMBER_INPUT_WIDGET":
			ret += "NUMBER_INPUT_WIDGET props be like " + COMPONENTS_BASE_PROMPT_NUMBER_INPUT_WIDGET + ". "
		case "SELECT_WIDGET":
			ret += "SELECT_WIDGET props be like " + COMPONENTS_BASE_PROMPT_SELECT_WIDGET + ". "
		case "CHART_WIDGET":
			ret += "CHART_WIDGET props be like " + COMPONENTS_BASE_PROMPT_CHART_WIDGET + ". "
		case "IMAGE_WIDGET":
			ret += "IMAGE_WIDGET props be like " + COMPONENTS_BASE_PROMPT_IMAGE_WIDGET + ". "
		case "UPLOAD_WIDGET":
			ret += "UPLOAD_WIDGET props be like " + COMPONENTS_BASE_PROMPT_UPLOAD_WIDGET + ". "
		case "EDITABLE_TEXT_WIDGET":
			ret += "EDITABLE_TEXT_WIDGET props be like " + COMPONENTS_BASE_PROMPT_EDITABLE_TEXT_WIDGET + ". "
		case "SLIDER_WIDGET":
			ret += "SLIDER_WIDGET props be like " + COMPONENTS_BASE_PROMPT_SLIDER_WIDGET + ". "
		case "RANGE_SLIDER_WIDGET":
			ret += "RANGE_SLIDER_WIDGET props be like " + COMPONENTS_BASE_PROMPT_RANGE_SLIDER_WIDGET + ". "
		case "SWITCH_WIDGET":
			ret += "SWITCH_WIDGET props be like " + COMPONENTS_BASE_PROMPT_SWITCH_WIDGET + ". "
		case "MULTISELECT_WIDGET":
			ret += "MULTISELECT_WIDGET props be like " + COMPONENTS_BASE_PROMPT_MULTISELECT_WIDGET + ". "
		case "CHECKBOX_GROUP_WIDGET":
			ret += "CHECKBOX_GROUP_WIDGET props be like " + COMPONENTS_BASE_PROMPT_CHECKBOX_GROUP_WIDGET + ". "
		}
	}
	egen.SaveHistoryMessageInRaw(ROLE_USER, ret)
	return ret
}

func (egen *EchoGenerator) FillContinueContext() {
	egen.SaveHistoryMessageInRaw(ROLE_USER, PRIMITIVE_PROMPT_CONTINUE)
}

func (egen *EchoGenerator) DoesContextIsContinued() bool {
	lastMessageSerial := len(egen.HistoryMessages) - 2
	if egen.HistoryMessages[lastMessageSerial].Content == PRIMITIVE_PROMPT_CONTINUE {
		return true
	}
	return false
}

func (egen *EchoGenerator) ConcatContinuedContextOnStackTop() {

	stackTopPointer := len(egen.HistoryMessages) - 1
	concatTargetPointer := stackTopPointer - 2

	fmt.Printf("[DUMP] ConcatContinuedContextOnStackTop\n")
	fmt.Printf("------------------------------------------------------------------\n\n")
	fmt.Printf("[DUMP] full[%d]: %+v\n", concatTargetPointer, egen.HistoryMessages[concatTargetPointer])
	fmt.Printf("[DUMP] full[%d]: %+v\n", stackTopPointer, egen.HistoryMessages[stackTopPointer])
	// concat
	egen.HistoryMessages[concatTargetPointer].Content += egen.HistoryMessages[stackTopPointer].Content
	// remove old context
	egen.HistoryMessages = egen.HistoryMessages[:concatTargetPointer+1]
}

// auto complete missing component field and properties
func (egen *EchoGenerator) ComponentFilter(uncompleteComponent string) string {
	return uncompleteComponent
}

func (egen *EchoGenerator) EmitEchoRequest(lastQueryDidNotFinish bool) (*HistoryMessage, error) {
	// emit new request
	tokenValidator := tokenvalidator.NewRequestTokenValidator()
	echoRequest := NewEchoRequest()
	fullHistoryMessage := egen.ExportFullHistoryMessages()
	echoRequest.SetMessages(fullHistoryMessage)
	echoPeripheralRequest := NewEchoPeripheralRequest(echoRequest.Export())
	token := tokenValidator.GenerateValidateToken(echoPeripheralRequest.Message)
	echoPeripheralRequest.SetValidateToken(token)

	// call API again
	echoFeedback, errInCallEcho := Echo(echoPeripheralRequest)
	if errInCallEcho != nil {
		return nil, errInCallEcho
	}
	fmt.Printf("[DUMP] EmitEchoRequest().echoFeedback: %+v\n", echoFeedback)

	// try export
	historyMessageFinal, queryDidNotFinish, errInExport := echoFeedback.ExportMessage()
	if errInExport != nil {
		return nil, errInExport
	}
	egen.SaveHistoryMessage(historyMessageFinal)

	// concat last not finished query
	if lastQueryDidNotFinish {
		egen.ConcatContinuedContextOnStackTop()
	}

	// query are finished
	if !queryDidNotFinish {
		return historyMessageFinal, nil
	}

	// query did not finish, request again
	egen.FillContinueContext()
	return egen.EmitEchoRequest(queryDidNotFinish)

}

type HistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (hm *HistoryMessage) UnMarshalContent() (map[string]interface{}, error) {
	// decode content
	var decodedContent map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(hm.Content), &decodedContent)
	if errInUnMarshal != nil {
		return nil, errInUnMarshal
	}
	return decodedContent, nil
}

func (hm *HistoryMessage) DetectComponentTypes() map[string]bool {
	component, _ := hm.UnMarshalContent()
	componentTypeList := make(map[string]bool)
	retrieveComponentTypes(component, componentTypeList)
	return componentTypeList
}

func retrieveComponentTypes(rawComponent map[string]interface{}, componentTypeList map[string]bool) {
	hitType, ok := rawComponent["type"]
	if !ok {
		return
	}
	hitTypeString, assertHitTypeOK := hitType.(string)
	if !assertHitTypeOK {
		return
	}
	componentTypeList[hitTypeString] = true
	hitChindrenNode, ok := rawComponent["childrenNode"]
	if !ok {
		return
	}
	hitChindrenNodeAsserted, asserthitChindrenNodeOK := hitChindrenNode.([]interface{})
	if !asserthitChindrenNodeOK {
		return
	}
	if len(hitChindrenNodeAsserted) == 0 {
		return
	}
	for _, node := range hitChindrenNodeAsserted {
		nodeAsserted, assertNodeOK := node.(map[string]interface{})
		if !assertNodeOK {
			continue
		}
		retrieveComponentTypes(nodeAsserted, componentTypeList)
	}
	return
}
