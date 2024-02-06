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
	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const SQL_SET_ACTION_PUBLIC = `update actions set config = jsonb_set(config, '{public}', 'true'::jsonb, true), updated_by= ? where team_id = ? and app_ref_id = ?;`
const SQL_SET_ACTION_PRIVATE = `update actions set config = jsonb_set(config, '{public}', 'false'::jsonb, true), updated_by= ? where team_id = ? and app_ref_id = ?;`

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

func (impl *ActionStorage) UpdateWholeAction(action *model.Action) error {
	if err := impl.db.Model(action).Where("id = ?", action.ID).UpdateColumns(action).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionStorage) UpdatePrivacyByTeamIDAndAppIDAndUserID(teamID int, appID int, userID int, isPublic bool) error {
	actions, errInGetAll := impl.RetrieveAll(teamID, appID)
	if errInGetAll != nil {
		return errInGetAll
	}
	// set status
	for _, action := range actions {
		if isPublic {
			action.SetPublic(userID)
		} else {
			action.SetPrivate(userID)
		}
		errInUpdateAction := impl.UpdateWholeAction(action)
		if errInUpdateAction != nil {
			return errInUpdateAction
		}
	}
	return nil
}

func (impl *ActionStorage) MakeActionPublicByTeamIDAndAppID(teamID int, appID int, userID int) error {
	tx := impl.db.Exec(SQL_SET_ACTION_PUBLIC, userID, teamID, appID)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (impl *ActionStorage) MakeActionPrivateByTeamIDAndAppID(teamID int, appID int, userID int) error {
	tx := impl.db.Exec(SQL_SET_ACTION_PRIVATE, userID, teamID, appID)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (impl *ActionStorage) RetrieveActionByTeamIDAndID(teamID int, actionID int) (*model.Action, error) {
	var action *model.Action
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, actionID).Find(&action).Error; err != nil {
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

func (impl *ActionStorage) RetrieveActionsByTeamIDAppIDVersionAndType(teamID int, appID int, version int, actionType int) ([]*model.Action, error) {
	var actions []*model.Action
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ? AND type = ?", teamID, appID, version, actionType).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *ActionStorage) RetrieveActionByTeamIDActionID(teamID int, actionID int) (*model.Action, error) {
	var action *model.Action
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, actionID).First(&action).Error; err != nil {
		return nil, err
	}
	return action, nil
}

func (impl *ActionStorage) DeleteActionsByApp(teamID int, appID int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Delete(&model.Action{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionStorage) DeleteActionByTeamIDAndActionID(teamID int, actionID int) error {
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, actionID).Delete(&model.Action{}).Error; err != nil {
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
