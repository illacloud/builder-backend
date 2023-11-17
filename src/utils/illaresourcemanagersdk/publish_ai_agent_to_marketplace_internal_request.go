package illaresourcemanagersdk

import "encoding/json"

type PublishAIAgentToMarketplaceInternalRequest struct {
	PublishedToMarketplace bool `json:"publishedToMarketplace"      `
	UserID                 int  `json:"userID"        validate:"required"`
}

func (req *PublishAIAgentToMarketplaceInternalRequest) ExportInJSONString() string {
	jsonByte, _ := json.Marshal(req)
	return string(jsonByte)
}

func NewPublishAIAgentToMarketplaceInternalRequest() *PublishAIAgentToMarketplaceInternalRequest {
	return &PublishAIAgentToMarketplaceInternalRequest{}
}

func NewPublishAIAgentToMarketplaceInternalRequestWithParam(publishedToMarketplace bool, userID int) *PublishAIAgentToMarketplaceInternalRequest {
	return &PublishAIAgentToMarketplaceInternalRequest{
		PublishedToMarketplace: publishedToMarketplace,
		UserID:                 userID,
	}
}
