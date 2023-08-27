package illamarketplacesdk

import "github.com/illacloud/builder-backend/src/model"

type AIAgentForMarketplace struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

func NewAIAgentForMarketplace(aiAgent *model.AIAgent) *AIAgentForMarketplace {
	aiAgentConfig := aiAgent.ExportConfig()
	aiAgentModelPayload := aiAgent.ExportModelPayload()
	return &AIAgentForMarketplace{
		Name:        aiAgent.Name,
		Description: aiAgentConfig.Description,
		Prompt:      aiAgentModelPayload.Prompt,
	}
}
