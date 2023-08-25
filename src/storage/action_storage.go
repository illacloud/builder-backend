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

package storage

import (
	"encoding/json"

	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ActionStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewActionStorage(logger *zap.SugaredLogger, db *gorm.DB) *ActionStorage {
	return &ActionStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *ActionStorage) Create(action *model.Action) (int, error) {
	if err := impl.db.Create(action).Error; err != nil {
		return 0, err
	}
	return action.ID, nil
}

func (impl *ActionStorage) Delete(teamID int, actionID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", actionID, teamID).Delete(&model.Action{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionStorage) Update(action *model.Action) error {
	if err := impl.db.Model(action).UpdateColumns(model.Action{
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

func (impl *ActionStorage) UpdateWholeAction(action *model.Action) error {
	if err := impl.db.Model(action).Where("id = ?", action.ID).UpdateColumns(action).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionStorage) UpdatePublicByTeamIDAndAppIDAndUserID(teamID int, appID int, userID int, actionConfig *model.ActionConfig) error {
	actions, errInGetAll := impl.RetrieveAll(teamID, appID)
	if errInGetAll != nil {
		return errInGetAll
	}
	// set status
	for _, action := range actions {
		tmpActionConfig := model.NewActionConfig()
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

func (impl *ActionStorage) MakeActionPublicByTeamIDAndAppID(teamID int, appID int, userID int) error {
	actions, errInGetAll := impl.RetrieveAll(teamID, appID)
	if errInGetAll != nil {
		return errInGetAll
	}
	// set status
	for _, action := range actions {
		action.SetPublic(userID)
		// update
		errorInUpdate := impl.UpdateWholeAction(action)
		if errorInUpdate != nil {
			return errorInUpdate
		}

	}
	return nil
}

func (impl *ActionStorage) MakeActionPrivateByTeamIDAndAppID(teamID int, appID int, userID int) error {
	actions, errInGetAll := impl.RetrieveAll(teamID, appID)
	if errInGetAll != nil {
		return errInGetAll
	}
	// set status
	for _, action := range actions {
		action.SetPrivate(userID)
		// update
		errorInUpdate := impl.UpdateWholeAction(action)
		if errorInUpdate != nil {
			return errorInUpdate
		}
	}
	return nil
}

func (impl *ActionStorage) RetrieveActionByIDAndTeamID(actionID int, teamID int) (*model.Action, error) {
	var action *model.Action
	if err := impl.db.Where("id = ? AND team_id = ?", actionID, teamID).Find(&action).Error; err != nil {
		return nil, err
	}
	return action, nil
}

func (impl *ActionStorage) RetrieveAll(teamID int, appID int) ([]*model.Action, error) {
	var actions []*model.Action
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *ActionStorage) RetrieveByID(teamID int, actionID int) (*model.Action, error) {
	var action *model.Action
	if err := impl.db.Where("id = ? AND team_id = ?", actionID, teamID).First(&action).Error; err != nil {
		return &model.Action{}, err
	}
	return action, nil
}

func (impl *ActionStorage) RetrieveActionsByTeamIDAppIDAndVersion(teamID int, appID int, version int) ([]*model.Action, error) {
	var actions []*model.Action
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, appID, version).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *ActionStorage) DeleteActionsByApp(teamID int, appID int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Delete(&model.Action{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionStorage) CountActionByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&model.Action{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *ActionStorage) DeleteAllActionsByTeamIDAppIDAndVersion(teamID int, appID int, targetVersion int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, appID, targetVersion).Delete(&model.Action{}).Error; err != nil {
		return err
	}
	return nil
}
