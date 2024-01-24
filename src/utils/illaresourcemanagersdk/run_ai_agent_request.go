package illaresourcemanagersdk

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/tokenvalidator"
)

const (
	RUN_AI_AGENT_REQUEST_FIELD_TEAM_ID          = "teamID"
	RUN_AI_AGENT_REQUEST_FIELD_RESOURCE_ID      = "resourceID"
	RUN_AI_AGENT_REQUEST_FIELD_AI_AGENT_ID      = "aiAgentID"
	RUN_AI_AGENT_REQUEST_FIELD_AUTHORIZATION    = "authorization"
	RUN_AI_AGENT_REQUEST_FIELD_RUN_BY_ANONYMOUS = "runByAnonymous"
	RUN_AI_AGENT_REQUEST_FIELD_AGENT_TYPE       = "agentType"
	RUN_AI_AGENT_REQUEST_FIELD_MODEL            = "model"
	RUN_AI_AGENT_REQUEST_FIELD_VARIABLES        = "variables"
	RUN_AI_AGENT_REQUEST_FIELD_MODEL_CONFIG     = "modelConfig"
	RUN_AI_AGENT_REQUEST_FIELD_INPUT            = "input"
)

// the request json like
//
//	{
//	    "agentType": 2,
//	    "model": 1,
//	    "variables":  [{"key":"key1", "value":"12"}, {"key":"key2", "value":"apple"}],
//	    "modelConfig": {"maxTokens": 100, "stream": false},
//	    "input": "can you tell me a story"
//	}
type RunAIAgentRequest struct {
	TeamID         int                      `json:"-"`
	AIAgentID      int                      `json:"-"` // alias of "resourceID"
	Authorization  string                   `json:"-"`
	RunByAnonymous bool                     `json:"runByAnonymous"`
	AgentType      int                      `json:"agentType"`
	Model          int                      `json:"model"`
	Variables      []*AIAgentPromptVariable `json:"variables"`
	ModelConfig    *AIAgentModelConfig      `json:"modelConfig"`
	Input          string                   `json:"input"`
}

func assertMagicStringIDToInt(magicStringID interface{}) (int, bool) {
	magicStringIDAsserted, assertPass := magicStringID.(string)
	if !assertPass {
		return 0, assertPass
	}
	return idconvertor.ConvertStringToInt(magicStringIDAsserted), true
}

func assertNumberIDToInt(floatID interface{}) (int, bool) {
	floatIDAsserted, assertPass := floatID.(float64)
	if assertPass {
		return int(floatIDAsserted), true

	}
	floatIDIntAsserted, assertPassInInt := floatID.(int)
	if assertPassInInt {
		return floatIDIntAsserted, true

	}
	return 0, false
}

func NewRunAIAgentRequest(rawRequest map[string]interface{}) (*RunAIAgentRequest, error) {
	runAIAgentRequest := &RunAIAgentRequest{}
	assertPass := true
	var variablesRaw []interface{}
	var variablesRawWarrped []map[string]interface{}
	var modelConfigRaw map[string]interface{}
	var errInNewModelConfig error
	for key, value := range rawRequest {
		switch key {
		case RUN_AI_AGENT_REQUEST_FIELD_TEAM_ID:
			// accept string & float64 type id
			runAIAgentRequest.TeamID, assertPass = assertMagicStringIDToInt(value)
			if !assertPass {
				runAIAgentRequest.TeamID, assertPass = assertNumberIDToInt(value)
			}
		case RUN_AI_AGENT_REQUEST_FIELD_RESOURCE_ID:
			// accept string & float64 type id
			runAIAgentRequest.AIAgentID, assertPass = assertMagicStringIDToInt(value)
			if !assertPass {
				runAIAgentRequest.AIAgentID, assertPass = assertNumberIDToInt(value)
			}
		case RUN_AI_AGENT_REQUEST_FIELD_AI_AGENT_ID:
			// accept string & float64 type id
			runAIAgentRequest.AIAgentID, assertPass = assertMagicStringIDToInt(value)
			if !assertPass {
				runAIAgentRequest.AIAgentID, assertPass = assertNumberIDToInt(value)
			}
		case RUN_AI_AGENT_REQUEST_FIELD_AUTHORIZATION:
			runAIAgentRequest.Authorization, assertPass = value.(string)
		case RUN_AI_AGENT_REQUEST_FIELD_RUN_BY_ANONYMOUS:
			runAIAgentRequest.RunByAnonymous, assertPass = value.(bool)
		case RUN_AI_AGENT_REQUEST_FIELD_AGENT_TYPE:
			runAIAgentRequest.AgentType, assertPass = assertNumberIDToInt(value)
		case RUN_AI_AGENT_REQUEST_FIELD_MODEL:
			runAIAgentRequest.Model, assertPass = assertNumberIDToInt(value)
		case RUN_AI_AGENT_REQUEST_FIELD_VARIABLES:
			variablesRaw, assertPass = value.([]interface{})
			if assertPass {
				for _, variableRaw := range variablesRaw {
					variableAsserted, assertPass := variableRaw.(map[string]interface{})
					if assertPass {
						variable, errInNewVariable := NewAIAgentPromptVariableByMap(variableAsserted)
						if errInNewVariable != nil {
							return nil, errInNewVariable
						}
						runAIAgentRequest.Variables = append(runAIAgentRequest.Variables, variable)
					} else {
						break
					}
				}
			} else {
				variablesRawWarrped, assertPass = value.([]map[string]interface{})
				if assertPass {
					for _, variableRaw := range variablesRawWarrped {
						variable, errInNewVariable := NewAIAgentPromptVariableByMap(variableRaw)
						if errInNewVariable != nil {
							return nil, errInNewVariable
						}
						runAIAgentRequest.Variables = append(runAIAgentRequest.Variables, variable)
					}
				}
			}
		case RUN_AI_AGENT_REQUEST_FIELD_MODEL_CONFIG:
			modelConfigRaw, assertPass = value.(map[string]interface{})
			if assertPass {
				runAIAgentRequest.ModelConfig, errInNewModelConfig = NewAIAgentModelConfigByMap(modelConfigRaw)
				if errInNewModelConfig != nil {
					return nil, errInNewModelConfig
				}
			}
		case RUN_AI_AGENT_REQUEST_FIELD_INPUT:
			runAIAgentRequest.Input, assertPass = value.(string)
		default:

		}
		if !assertPass {
			return nil, errors.New(fmt.Sprintf("assert request field '%s' with value '%s' for RunAIAgentRequest", key, value))
		}
	}
	return runAIAgentRequest, nil
}

func (req *RunAIAgentRequest) ExportTeamIDInMagicString() string {
	return idconvertor.ConvertIntToString(req.TeamID)
}

func (req *RunAIAgentRequest) ExportTeamIDInInt() int {
	return req.TeamID
}

func (req *RunAIAgentRequest) ExportAIAgentIDInMagicString() string {
	return idconvertor.ConvertIntToString(req.AIAgentID)
}

func (req *RunAIAgentRequest) ExportAIAgentIDInInt() int {
	return req.AIAgentID
}

func (req *RunAIAgentRequest) ExportAuthorization() string {
	return req.Authorization
}

func (req *RunAIAgentRequest) IsRunByAnonymous() bool {
	return req.RunByAnonymous
}

func (req *RunAIAgentRequest) ExportRequestToken() string {
	tokenValidator := tokenvalidator.NewRequestTokenValidator()
	return tokenValidator.GenerateValidateToken(strconv.Itoa(req.AIAgentID))
}
