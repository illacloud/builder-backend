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
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/util/resourcelist"
	"github.com/illacloud/builder-backend/pkg/db"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ActionRepository interface {
	Create(action *Action) (int, error)
	Delete(teamID int, actionID int) error
	Update(action *Action) error
	UpdatePublicByTeamIDAndAppIDAndUserID(teamID int, appID int, userID int, actionConfig *ActionConfig) error
	RetrieveActionByIDAndTeamID(actionID int, teamID int) (*Action, error)
	RetrieveAll(teamID int, appID int) ([]*Action, error)
	RetrieveByID(teamID int, actionID int) (*Action, error)
	RetrieveActionsByAppVersion(teamID int, appID int, version int) ([]*Action, error)
	DeleteActionsByApp(teamID int, appID int) error
	CountActionByTeamID(teamID int) (int, error)
}

type ActionRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewActionRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *ActionRepositoryImpl {
	return &ActionRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *ActionRepositoryImpl) Create(action *Action) (int, error) {
	if err := impl.db.Create(action).Error; err != nil {
		return 0, err
	}
	return action.ID, nil
}

func (impl *ActionRepositoryImpl) Delete(teamID int, actionID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", actionID, teamID).Delete(&Action{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) Update(action *Action) error {
	if err := impl.db.Model(action).UpdateColumns(Action{
		Resource:    action.Resource,
		Type:        action.Type,
		Name:        action.Name,
		TriggerMode: action.TriggerMode,
		Transformer: action.Transformer,
		Template:    action.Template,
		Config:      action.Config,
		UpdatedBy:   action.UpdatedBy,
		UpdatedAt:   action.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) UpdatePublicByTeamIDAndAppIDAndUserID(teamID int, appID int, userID int, actionConfig *ActionConfig) error {
	actions, errInGetAll := impl.RetrieveAll(teamID, appID)
	if errInGetAll != nil {
		return errInGetAll
	}
	// set status
	for _, action := range actions {
		tmpActionConfig := NewActionConfig()
		json.Unmarshal([]byte(action.Config), &tmpActionConfig)
		tmpActionConfig.Public = actionConfig.Public
		action.Config = tmpActionConfig.ExportToJSONString()
		// update
		errorInUpdate := impl.Update(action)
		if errorInUpdate != nil {
			return errorInUpdate
		}

	}
	return nil
}

func (impl *ActionRepositoryImpl) RetrieveActionByIDAndTeamID(actionID int, teamID int) (*Action, error) {
	var action *Action
	if err := impl.db.Where("id = ? AND team_id = ?", actionID, teamID).Find(&action).Error; err != nil {
		return nil, err
	}
	return action, nil
}

func (impl *ActionRepositoryImpl) RetrieveAll(teamID int, appID int) ([]*Action, error) {
	var actions []*Action
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *ActionRepositoryImpl) RetrieveByID(teamID int, actionID int) (*Action, error) {
	var action *Action
	if err := impl.db.Where("id = ? AND team_id = ?", actionID, teamID).First(&action).Error; err != nil {
		return &Action{}, err
	}
	return action, nil
}

func (impl *ActionRepositoryImpl) RetrieveActionsByAppVersion(teamID int, appID int, version int) ([]*Action, error) {
	var actions []*Action
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, appID, version).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *ActionRepositoryImpl) DeleteActionsByApp(teamID int, appID int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Delete(&Action{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) CountActionByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&Action{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
