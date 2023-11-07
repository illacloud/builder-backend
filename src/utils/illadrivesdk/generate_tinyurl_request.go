package illadrivesdk

import (
	"time"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

type GenerateTinyURLRequest struct {
	IDs               []string `json:"ids" validate:"required,gt=0,dive,required"`
	IntIDs            []int    `json:"-"`
	ExpirationType    string   `json:"expirationType" validate:"oneof=persistent custom"`
	Expiry            string   `json:"expiry" validate:"required_if=ExpirationType custom"`
	HotlinkProtection bool     `json:"hotlinkProtection"`
}

func NewGenerateTinyURLRequest() *GenerateTinyURLRequest {
	return &GenerateTinyURLRequest{}
}

func NewGenerateTinyURLRequestByParam(ids []string, expirationType string, expiry string, hotlinkProtection bool) *GenerateTinyURLRequest {
	return &GenerateTinyURLRequest{
		IDs:               ids,
		ExpirationType:    expirationType,
		Expiry:            expiry,
		HotlinkProtection: hotlinkProtection,
	}
}

func (g *GenerateTinyURLRequest) Preprocess() {
	for _, id := range g.IDs {
		g.IntIDs = append(g.IntIDs, idconvertor.ConvertStringToInt(id))
	}
}

func (g *GenerateTinyURLRequest) ExportIDs() []int {
	return g.IntIDs
}

func (g *GenerateTinyURLRequest) ExportExpirationType() int {
	if g.ExpirationType == "custom" {
		return 2
	}
	return 1
}

func (g *GenerateTinyURLRequest) ValidateExpiry() bool {
	if g.ExpirationType == "custom" {
		if g.Expiry == "" {
			return false
		} else {
			if _, err := time.ParseDuration(g.Expiry); err != nil {
				return false
			}
		}
	}
	return true
}
