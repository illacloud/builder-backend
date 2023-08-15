package illaresourcemanagerbackendsdk

const (
	RUN_AI_AGENT_RESULT_FIELD_PAYLOAD = "payload"
)

// the result json like
//
// {
// "payload": "Once upon a time in the beautiful village of Fruitland, there lived a young girl named Apple. Apple was known for her vibrant red hair that matched the color of the apples that grew abundantly in her family's orchard. She was full of curiosity and had a heart full of kindness.\n\nApple had always dreamt of exploring the world beyond Fruitland. She had heard countless stories from her grandparents about faraway lands filled with magical creatures and stunning landscapes. Yearning for adventure, she embarked on"
// }
type RunAIAgentResult struct {
	Payload string `json:"payload"`
}

func NewRunAIAgentResult() *RunAIAgentResult {
	return &RunAIAgentResult{}
}

func (i *RunAIAgentResult) ExportPayload() string {
	return i.Payload
}
