package illaresourcemanagersdk

import (
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

type AIAgentForExport struct {
	ID                     string                 `json:"aiAgentID"`
	UID                    uuid.UUID              `json:"uid"`
	TeamID                 string                 `json:"teamID"`
	TeamName               string                 `json:"teamName"`
	TeamIcon               string                 `json:"teamIcon"`
	TeamIdentifier         string                 `json:"teamIdentifier"`
	Name                   string                 `json:"name"`
	Model                  int                    `json:"model"`
	AgentType              int                    `json:"agentType"`
	PublishedToMarketplace bool                   `json:"publishedToMarketplace"`
	Icon                   string                 `json:"icon"`
	Description            string                 `json:"description"`
	Prompt                 string                 `json:"prompt"`
	Variables              []interface{}          `json:"variables"`
	ModelConfig            map[string]interface{} `json:"modelConfig"`
	CreatedAt              time.Time              `json:"createdAt"`
	CreatedBy              string                 `json:"createdBy"`
	UpdatedAt              time.Time              `json:"updatedAt"`
	UpdatedBy              string                 `json:"updatedBy"`
	EditedBy               []interface{}          `json:"editedBy"`
}

func (i *AIAgentForExport) ExportIDInInt() int {
	return idconvertor.ConvertStringToInt(i.ID)
}
