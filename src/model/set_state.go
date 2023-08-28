package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	SET_STATE_FIELD_DISPLAY_NAME = "displayName"
)

type SetState struct {
	ID        int       `json:"id" 		   gorm:"column:id;type:bigserial"`
	UID       uuid.UUID `json:"uid" 	   gorm:"column:uid;type:uuid;not null"`
	TeamID    int       `json:"teamID"    gorm:"column:team_id;type:bigserial"`
	StateType int       `json:"state_type" gorm:"column:state_type;type:bigint"`
	AppRefID  int       `json:"app_ref_id" gorm:"column:app_ref_id;type:bigint"`
	Version   int       `json:"version"    gorm:"column:version;type:bigint"`
	Value     string    `json:"value" 	   gorm:"column:value;type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamp"`
	CreatedBy int       `json:"created_by" gorm:"column:created_by;type:bigint"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	UpdatedBy int       `json:"updated_by" gorm:"column:updated_by;type:bigint"`
}

func NewSetStateByApp(app *App, stateType int) *SetState {
	setState := &SetState{
		TeamID:    app.ExportTeamID(),
		StateType: stateType,
		AppRefID:  app.ExportID(),
		Version:   APP_EDIT_VERSION,
		CreatedBy: app.ExportUpdatedBy(),
		UpdatedBy: app.ExportUpdatedBy(),
	}
	setState.InitUID()
	setState.InitCreatedAt()
	setState.InitUpdatedAt()
	return setState
}

func NewSetStateByWebsocketMessage(app *App, stateType int, data interface{}) (*SetState, error) {
	setState := NewSetStateByApp(app, stateType)
	udata, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("SetStateDto ConstructByMap failed, please check your input.")
	}
	displayName, mapok := udata[SET_STATE_FIELD_DISPLAY_NAME].(string)
	if !mapok {
		return nil, errors.New("SetStateDto ConstructByMap failed, can not find displayName field.")
	}
	// fild
	setState.Value = displayName
	return setState, nil
}

func (setState *SetState) CleanID() {
	setState.ID = 0
}

func (setState *SetState) InitUID() {
	setState.UID = uuid.New()
}

func (setState *SetState) InitCreatedAt() {
	setState.CreatedAt = time.Now().UTC()
}

func (setState *SetState) InitUpdatedAt() {
	setState.UpdatedAt = time.Now().UTC()
}

func (setState *SetState) InitForFork(teamID int, appID int, version int, userID int) {
	setState.TeamID = teamID
	setState.AppRefID = appID
	setState.Version = version
	setState.CreatedBy = userID
	setState.UpdatedBy = userID
	setState.CleanID()
	setState.InitUID()
	setState.InitCreatedAt()
	setState.InitUpdatedAt()
}

func (setState *SetState) AppendNewVersion(newVersion int) {
	setState.CleanID()
	setState.InitUID()
	setState.Version = newVersion
}

func (setState *SetState) UpdateByNewSetState(newSetState *SetState) {
	setState.Value = newSetState.Value
	setState.UpdatedBy = newSetState.UpdatedBy
	setState.InitUpdatedAt()
}

func (setState *SetState) ExportID() int {
	return setState.ID
}

func (setState *SetState) ExportValue() string {
	return setState.Value
}

func (setState *SetState) ExportVersion() int {
	return setState.Version
}
