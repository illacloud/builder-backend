package repository

import (
	"encoding/json"
	"errors"
)

type EchoFeedback struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Usage   map[string]interface{} `json:"usage"`
	Choices []interface{}          `json:"choices"`
}

func (ef *EchoFeedback) Avaliable() bool {
	if len(ef.Choices) == 0 {
		return false
	}
	return true
}

func (ef *EchoFeedback) ExportMessage() (*HistoryMessage, bool, error) {
	// init
	needQueryMore := false
	// pick up first choice
	if len(ef.Choices) == 0 {
		return nil, needQueryMore, errors.New("echo return 0 choices.")
	}
	firstChoice := ef.Choices[0]
	firstChoiceAsserted, assertFirstChoiceOK := firstChoice.(map[string]interface{})
	if !assertFirstChoiceOK {
		return nil, needQueryMore, errors.New("choices syntax illegal.")
	}
	// check out meessage is finish
	finishReason, ok := firstChoiceAsserted["finish_reason"]
	if !ok {
		return nil, needQueryMore, errors.New("choices finish_reason syntax illegal.")
	}
	if finishReason != "stop" {
		needQueryMore = true
	}
	// assert message
	message, ok := firstChoiceAsserted["message"]
	if !ok {
		return nil, needQueryMore, errors.New("choices message syntax illegal.")
	}
	messageAsserted, assertMessageOK := message.(map[string]interface{})
	if !assertMessageOK {
		return nil, needQueryMore, errors.New("message syntax illegal.")
	}
	// fill history message
	role, ok := messageAsserted["role"]
	if !ok {
		return nil, needQueryMore, errors.New("message.role syntax illegal.")
	}
	roleAsserted, assertRoleOK := role.(string)
	if !assertRoleOK {
		return nil, needQueryMore, errors.New("message.role assert failed.")
	}
	content, ok := messageAsserted["content"]
	if !ok {
		return nil, needQueryMore, errors.New("message.content syntax illegal.")
	}
	contentAsserted, assertContentOK := content.(string)
	if !assertContentOK {
		return nil, needQueryMore, errors.New("message.content assert failed.")
	}

	historyMessage := &HistoryMessage{
		Role:    roleAsserted,
		Content: contentAsserted,
	}
	return historyMessage, needQueryMore, nil
}

func (ef *EchoFeedback) ExportContent() (map[string]interface{}, bool, error) {
	// init
	needQueryMore := false
	// pick up first choice
	if len(ef.Choices) == 0 {
		return nil, needQueryMore, errors.New("echo return 0 choices.")
	}
	firstChoice := ef.Choices[0]
	firstChoiceAsserted, assertFirstChoiceOK := firstChoice.(map[string]interface{})
	if !assertFirstChoiceOK {
		return nil, needQueryMore, errors.New("choices syntax illegal.")
	}
	// check out meessage is finish
	finishReason, ok := firstChoiceAsserted["finish_reason"]
	if !ok {
		return nil, needQueryMore, errors.New("choices finish_reason syntax illegal.")
	}
	if finishReason != "stop" {
		needQueryMore = true
	}
	// assert message
	message, ok := firstChoiceAsserted["message"]
	if !ok {
		return nil, needQueryMore, errors.New("choices message syntax illegal.")
	}
	messageAsserted, assertMessageOK := message.(map[string]interface{})
	if !assertMessageOK {
		return nil, needQueryMore, errors.New("message syntax illegal.")
	}
	// assert content
	content, ok := messageAsserted["content"]
	if !ok {
		return nil, needQueryMore, errors.New("message content syntax illegal.")
	}
	contentAsserted, assertContentOK := content.(string)
	if !assertContentOK {
		return nil, needQueryMore, errors.New("content syntax illegal.")
	}
	// decode content
	var decodedContent map[string]interface{}
	errInUnMarshal := json.Unmarshal([]byte(contentAsserted), &decodedContent)
	if errInUnMarshal != nil {
		return nil, needQueryMore, errInUnMarshal
	}
	return decodedContent, needQueryMore, nil
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Message      map[string]interface{} `json:"message"`
	FinishReason string                 `json:"finish_reason"`
	Index        int                    `json:"index"`
}