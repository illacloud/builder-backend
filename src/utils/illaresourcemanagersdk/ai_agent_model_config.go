package illaresourcemanagersdk

import (
	"fmt"
)

const (
	AI_AGENT_MODEL_CONFIG_FIELD_SUFFIX            = "suffix"
	AI_AGENT_MODEL_CONFIG_FIELD_MAX_TOKENS        = "maxTokens"
	AI_AGENT_MODEL_CONFIG_FIELD_TEMPERATURE       = "temperature"
	AI_AGENT_MODEL_CONFIG_FIELD_TOP_P             = "topP"
	AI_AGENT_MODEL_CONFIG_FIELD_N                 = "n"
	AI_AGENT_MODEL_CONFIG_FIELD_STREAM            = "stream"
	AI_AGENT_MODEL_CONFIG_FIELD_LOGPROBS          = "logprobs"
	AI_AGENT_MODEL_CONFIG_FIELD_ECHO              = "echo"
	AI_AGENT_MODEL_CONFIG_FIELD_STOP              = "stop"
	AI_AGENT_MODEL_CONFIG_FIELD_PRESENCE_PENALTY  = "presencePenalty"
	AI_AGENT_MODEL_CONFIG_FIELD_FREQUENCY_PENALTY = "frequencyPenalty"
	AI_AGENT_MODEL_CONFIG_FIELD_BEST_OF           = "bestOf"
	AI_AGENT_MODEL_CONFIG_FIELD_LOGIT_BIAS        = "logitBias"
	AI_AGENT_MODEL_CONFIG_FIELD_USER              = "user"
)

type AIAgentModelConfigSettedMap struct {
	Suffix           bool
	MaxTokens        bool
	Temperature      bool
	TopP             bool
	N                bool
	Stream           bool
	Logprobs         bool
	Echo             bool
	Stop             bool
	PresencePenalty  bool
	FrequencyPenalty bool
	BestOf           bool
	LogitBias        bool
	User             bool
}

type AIAgentModelConfig struct {
	Suffix           string                 `json:"suffix"`            // default is null
	MaxTokens        int                    `json:"maxTokens"`         // default is 16
	Temperature      float64                `json:"temperature"`       // default is 1
	TopP             float64                `json:"top_p"`             // default is 1
	N                float64                `json:"n"`                 // default is 1
	Stream           bool                   `json:"stream"`            // default is false
	Logprobs         float64                `json:"logprobs"`          // default is 0
	Echo             bool                   `json:"echo"`              // default is false
	Stop             []string               `jsonstop"`                // default is null
	PresencePenalty  float64                `json:"presence_penalty"`  // default is 0
	FrequencyPenalty float64                `json:"frequency_penalty"` // default is 0
	BestOf           float64                `json:"best_of"`           // default is 1
	LogitBias        map[string]interface{} `json:"logit_bias"`        // default is null
	User             string                 `json:"user"`              // default is null
	SettedMap        *AIAgentModelConfigSettedMap
}

// only reflect field: maxTokens, temperature, stream
func ReflectAIAgentModelConfigManually(aiAgentModelConfig *AIAgentModelConfig, rawRequest map[string]interface{}) (*AIAgentModelConfig, error) {
	assertPass := true
	for key, value := range rawRequest {
		switch key {
		case AI_AGENT_MODEL_CONFIG_FIELD_SUFFIX:
			aiAgentModelConfig.Suffix, assertPass = value.(string)
		case AI_AGENT_MODEL_CONFIG_FIELD_MAX_TOKENS:
			var maxTokenFloat float64
			maxTokenFloat, assertPass = value.(float64)
			aiAgentModelConfig.MaxTokens = int(maxTokenFloat)
			aiAgentModelConfig.SettedMap.MaxTokens = true
		case AI_AGENT_MODEL_CONFIG_FIELD_TEMPERATURE:
			aiAgentModelConfig.Temperature, assertPass = value.(float64)
			aiAgentModelConfig.SettedMap.Temperature = true
		case AI_AGENT_MODEL_CONFIG_FIELD_TOP_P:
			aiAgentModelConfig.TopP, assertPass = value.(float64)
		case AI_AGENT_MODEL_CONFIG_FIELD_N:
			aiAgentModelConfig.N, assertPass = value.(float64)
		case AI_AGENT_MODEL_CONFIG_FIELD_STREAM:
			aiAgentModelConfig.Stream, assertPass = value.(bool)
			aiAgentModelConfig.SettedMap.Stream = true
		case AI_AGENT_MODEL_CONFIG_FIELD_LOGPROBS:
			aiAgentModelConfig.Logprobs, assertPass = value.(float64)
		case AI_AGENT_MODEL_CONFIG_FIELD_ECHO:
			aiAgentModelConfig.Echo, assertPass = value.(bool)
		case AI_AGENT_MODEL_CONFIG_FIELD_STOP:
			aiAgentModelConfig.Stop, assertPass = value.([]string)
		case AI_AGENT_MODEL_CONFIG_FIELD_PRESENCE_PENALTY:
			aiAgentModelConfig.PresencePenalty, assertPass = value.(float64)
		case AI_AGENT_MODEL_CONFIG_FIELD_FREQUENCY_PENALTY:
			aiAgentModelConfig.FrequencyPenalty, assertPass = value.(float64)
		case AI_AGENT_MODEL_CONFIG_FIELD_BEST_OF:
			aiAgentModelConfig.BestOf, assertPass = value.(float64)
		case AI_AGENT_MODEL_CONFIG_FIELD_LOGIT_BIAS:
			aiAgentModelConfig.LogitBias, assertPass = value.(map[string]interface{})
		case AI_AGENT_MODEL_CONFIG_FIELD_USER:
			aiAgentModelConfig.User, assertPass = value.(string)
		default:
		}
		if !assertPass {
			fmt.Printf("[ReflectAIAgentModelConfigManually] assertFailed\n")
		}
	}
	return aiAgentModelConfig, nil
}

func NewAIAgentModelConfigByMap(rawRequest map[string]interface{}) (*AIAgentModelConfig, error) {
	settedMap := &AIAgentModelConfigSettedMap{
		Stream:      true,
		Temperature: true,
		MaxTokens:   true,
	}
	aiAgentModelConfig := &AIAgentModelConfig{
		Stream:      false,
		Temperature: 1.0,
		MaxTokens:   1024,
		SettedMap:   settedMap,
	}
	return ReflectAIAgentModelConfigManually(aiAgentModelConfig, rawRequest)

}
