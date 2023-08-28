package illaresourcemanagersdk

import (
	"errors"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const (
	RUN_AI_AGENT_REQUEST_FIELD_TEAM_ID          = "teamID"
	RUN_AI_AGENT_REQUEST_FIELD_RESOURCE_ID      = "resourceID"
	RUN_AI_AGENT_REQUEST_FIELD_AIAGENT_ID       = "aiAgentID"
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
	TeamID         string                   `json:"-"`
	AIAgentID      string                   `json:"-"` // alias of "resourceID"
	Authorization  string                   `json:"-"`
	RunByAnonymous bool                     `json:"runByAnonymous"`
	AgentType      int                      `json:"agentType"`
	Model          int                      `json:"model"`
	Variables      []*AIAgentPromptVariable `json:"variables"`
	ModelConfig    *AIAgentModelConfig      `json:"modelConfig"`
	Input          string                   `json:"input"`
}

func NewRunAIAgentRequest(rawRequest map[string]interface{}) (*RunAIAgentRequest, error) {
	runAIAgentRequest := &RunAIAgentRequest{}
	assertPass := true
	var agentTypeFloat64 float64
	var modelFloat64 float64
	var variablesRaw []interface{}
	var modelConfigRaw map[string]interface{}
	var errInNewModelConfig error
	for key, value := range rawRequest {
		switch key {
		case RUN_AI_AGENT_REQUEST_FIELD_TEAM_ID:
			runAIAgentRequest.TeamID, assertPass = value.(string)
		case RUN_AI_AGENT_REQUEST_FIELD_RESOURCE_ID:
			runAIAgentRequest.AIAgentID, assertPass = value.(string)
		case RUN_AI_AGENT_REQUEST_FIELD_AIAGENT_ID:
			runAIAgentRequest.AIAgentID, assertPass = value.(string)
		case RUN_AI_AGENT_REQUEST_FIELD_AUTHORIZATION:
			runAIAgentRequest.Authorization, assertPass = value.(string)
		case RUN_AI_AGENT_REQUEST_FIELD_RUN_BY_ANONYMOUS:
			runAIAgentRequest.RunByAnonymous, assertPass = value.(bool)
		case RUN_AI_AGENT_REQUEST_FIELD_AGENT_TYPE:
			agentTypeFloat64, assertPass = value.(float64)
			if assertPass {
				runAIAgentRequest.AgentType = int(agentTypeFloat64)
			}
		case RUN_AI_AGENT_REQUEST_FIELD_MODEL:
			modelFloat64, assertPass = value.(float64)
			if assertPass {
				runAIAgentRequest.Model = int(modelFloat64)
			}
		case RUN_AI_AGENT_REQUEST_FIELD_VARIABLES:
			variablesRaw, assertPass = value.([]interface{})
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
			return nil, errors.New("assert request field failed for RunAIAgentRequest")
		}
	}
	return runAIAgentRequest, nil
}

func (req *RunAIAgentRequest) ExportTeamID() string {
	return req.TeamID
}

func (req *RunAIAgentRequest) ExportTeamIDInInt() int {
	return idconvertor.ConvertStringToInt(req.TeamID)
}

func (req *RunAIAgentRequest) ExportAIAgentID() string {
	return req.AIAgentID
}

func (req *RunAIAgentRequest) ExportAIAgentIDInInt() int {
	return idconvertor.ConvertStringToInt(req.AIAgentID)
}

func (req *RunAIAgentRequest) ExportAuthorization() string {
	return req.Authorization
}

func (req *RunAIAgentRequest) IsRunByAnonymous() bool {
	return req.RunByAnonymous
}
