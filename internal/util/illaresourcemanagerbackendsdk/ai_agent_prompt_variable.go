package illaresourcemanagerbackendsdk

import (
	"errors"
)

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
	var assertPass bool
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
			return nil, errors.New("new ai agent prompt variables failed due to assert failed.")
		}
	}
	return aiAgentPromptVariable, nil
}
