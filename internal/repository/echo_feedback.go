package repository

type EchoFeedback struct {
	Id      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Usage   map[string]interface{} `json:"usage"`
	Choices []interface{}          `json:"choices"`
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
