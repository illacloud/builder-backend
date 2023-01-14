// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

type SetStateRepository interface {
	Create(setState *SetState) error
	Delete(teamID int, setStateID int) error
	DeleteByValue(setState *SetState) error
	Update(setState *SetState) error
	UpdateByValue(beforeSetState *SetState, afterSetState *SetState) error
	RetrieveByID(teamID int, setStateID int) (*SetState, error)
	RetrieveSetStatesByVersion(teamID int, version int) ([]*SetState, error)
	RetrieveByValue(setState *SetState) (*SetState, error)
	RetrieveSetStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*SetState, error)
	DeleteAllTypeSetStatesByApp(teamID int, apprefid int) error
}

type SetStateRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewSetStateRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *SetStateRepositoryImpl {
	return &SetStateRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *SetStateRepositoryImpl) Create(setState *SetState) error {
	if err := impl.db.Create(setState).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateRepositoryImpl) Delete(teamID int, setStateID int) error {
	if err := impl.db.Delete(&SetState{}).Where("id = ? AND team_id = ?", setStateID, teamID).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateRepositoryImpl) DeleteByValue(setState *SetState) error {
	if err := impl.db.Where("team_id = ? AND value = ?", setState.TeamID, setState.Value).Delete(&SetState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateRepositoryImpl) Update(setState *SetState) error {
	if err := impl.db.Model(setState).UpdateColumns(SetState{
		ID:        setState.ID,
		StateType: setState.StateType,
		AppRefID:  setState.AppRefID,
		Version:   setState.Version,
		Value:     setState.Value,
		UpdatedAt: setState.UpdatedAt,
		UpdatedBy: setState.UpdatedBy,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateRepositoryImpl) UpdateByValue(beforeSetState *SetState, afterSetState *SetState) error {
	if err := impl.db.Model(afterSetState).Where(
		"app_ref_id = ? AND state_type = ? AND version = ? AND value = ?",
		beforeSetState.AppRefID,
		beforeSetState.StateType,
		beforeSetState.Version,
		beforeSetState.Value,
	).UpdateColumns(afterSetState).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateRepositoryImpl) RetrieveByID(teamID int, setStateID int) (*SetState, error) {
	var setState *SetState
	if err := impl.db.Where("team_id = ? AND value = ?", teamID, setState.Value).First(&setState).Error; err != nil {
		return &SetState{}, err
	}
	return setState, nil
}

func (impl *SetStateRepositoryImpl) RetrieveSetStatesByVersion(teamID int, version int) ([]*SetState, error) {
	var setStates []*SetState
	if err := impl.db.Where("team_id = ? AND version = ?", teamID, version).Find(&setStates).Error; err != nil {
		return nil, err
	}
	return setStates, nil
}

func (impl *SetStateRepositoryImpl) RetrieveByValue(setState *SetState) (*SetState, error) {
	var ret *SetState
	if err := impl.db.Where(
		"team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ? AND value = ?",
		setState.TeamID,
		setState.AppRefID,
		setState.StateType,
		setState.Version,
		setState.Value,
	).First(&ret).Error; err != nil {
		return nil, err
	}
	return ret, nil
}

func (impl *SetStateRepositoryImpl) RetrieveSetStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*SetState, error) {
	var setStates []*SetState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ?", teamID, apprefid, statetype, version).Find(&setStates).Error; err != nil {
		return nil, err
	}
	return setStates, nil
}

func (impl *SetStateRepositoryImpl) DeleteAllTypeSetStatesByApp(teamID int, apprefid int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, apprefid).Delete(&SetState{}).Error; err != nil {
		return err
	}
	return nil
}
