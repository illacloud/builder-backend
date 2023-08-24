package illacapacity

import (
	"time"

	"github.com/google/uuid"
)

type StartTransactionResponse struct {
	TeamID    int       `json:"teamID"`
	UID       uuid.UUID `json:"transactionUID"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func NewStartTransactionResponse() *StartTransactionResponse {
	return &StartTransactionResponse{}
}

func (resp *StartTransactionResponse) ExportTeamID() int {
	return resp.TeamID
}

func (resp *StartTransactionResponse) ExportTXUID() uuid.UUID {
	return resp.UID
}

func (resp *StartTransactionResponse) ExportTXUIDInString() string {
	return resp.UID.String()
}
