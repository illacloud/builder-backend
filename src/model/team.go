package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/illa-builder-backend/src/utils/idconvertor"
)

type RawTeam struct {
	ID         string    `json:"id"`
	UID        uuid.UUID `json:"uid"`
	Name       string    `json:"name"`
	Identifier string    `json:"identifier"`
	Icon       string    `json:"icon"`
	Permission string    `json:"permission"` // for team permission config
	CreatedAt  time.Time ``
	UpdatedAt  time.Time ``
}

type Team struct {
	ID         int       `json:"id"`
	UID        uuid.UUID `json:"uid"`
	Name       string    `json:"name"`
	Identifier string    `json:"identifier"`
	Icon       string    `json:"icon"`
	Permission string    `json:"permission"` // for team permission config
	CreatedAt  time.Time ``
	UpdatedAt  time.Time ``
}

func NewTeam(u *RawTeam) *Team {
	return &Team{
		ID:         idconvertor.ConvertStringToInt(u.ID),
		UID:        u.UID,
		Name:       u.Name,
		Identifier: u.Identifier,
		Icon:       u.Icon,
		Permission: u.Permission,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func NewTeamByDataControlRawData(rawTeamString string) (*Team, error) {
	rawTeam := RawTeam{}
	errInUnmarshal := json.Unmarshal([]byte(rawTeamString), &rawTeam)
	if errInUnmarshal != nil {
		return nil, errInUnmarshal
	}
	return NewTeam(&rawTeam), nil
}

func (u *Team) GetID() int {
	return u.ID
}
