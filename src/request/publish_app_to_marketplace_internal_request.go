package request

import "encoding/json"

type PublishAppToMarketplaceInternalRequest struct {
	PublishedToMarketplace bool `json:"publishedToMarketplace"      `
	UserID                 int  `json:"userID"        validate:"required"`
}

func (req *PublishAppToMarketplaceInternalRequest) ExportInJSONString() string {
	jsonByte, _ := json.Marshal(req)
	return string(jsonByte)
}

func NewPublishAppToMarketplaceInternalRequest() *PublishAppToMarketplaceInternalRequest {
	return &PublishAppToMarketplaceInternalRequest{}
}
