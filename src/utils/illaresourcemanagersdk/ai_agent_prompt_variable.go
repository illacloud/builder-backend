package illaresourcemanagersdk

import "fmt"

const (
	AIAGENTPROMPTVARIABLE_FIELD_KEY           = "key"
	AIAGENTPROMPTVARIABLE_FIELD_VALUE         = "value"
	AIAGENTPROMPTVARIABLE_FIELD_DEFAULT_VALUE = "defaultValue"
)

type AIAgentPromptVariable struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	DefaultValue string `json:"defaultValue"`
}

func NewAIAgentPromptVariableByMap(rawData map[string]interface{}) (*AIAgentPromptVariable, error) {
	aiAgentPromptVariable := &AIAgentPromptVariable{}
	assertPass := true
	for key, value := range rawData {
		switch key {
		case AIAGENTPROMPTVARIABLE_FIELD_KEY:
			aiAgentPromptVariable.Key, assertPass = value.(string)
		case AIAGENTPROMPTVARIABLE_FIELD_VALUE:
			aiAgentPromptVariable.Value, assertPass = value.(string)
		case AIAGENTPROMPTVARIABLE_FIELD_DEFAULT_VALUE:
			aiAgentPromptVariable.DefaultValue, assertPass = value.(string)
		}
		if !assertPass {
			return nil, fmt.Errorf("new ai agent prompt variables failed due to assert failed with key: '%s', value: '%+v'\n", key, value)
		}
	}
	return aiAgentPromptVariable, nil
}
