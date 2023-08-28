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

type KVStateStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewKVStateStorage(logger *zap.SugaredLogger, db *gorm.DB) *KVStateStorage {
	return &KVStateStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *KVStateStorage) Create(kvState *model.KVState) error {
	if err := impl.db.Create(kvState).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) Delete(teamID int, kvStateID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", kvStateID, teamID).Delete(&model.KVState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) Update(kvState *model.KVState) error {
	if err := impl.db.Model(kvState).Where("id = ?", kvState.ID).UpdateColumns(kvState).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) RetrieveByID(teamID int, kvStateID int) (*model.KVState, error) {
	var kvState *model.KVState
	if err := impl.db.Where("id = ? AND team_id = ?", kvStateID, teamID).First(&kvState).Error; err != nil {
		return &model.KVState{}, err
	}
	return kvState, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByVersion(teamID int, version int) ([]*model.KVState, error) {
	var kvStates []*model.KVState
	if err := impl.db.Where("team_id = ? AND version = ?", teamID, version).Find(&kvStates).Error; err != nil {
		return nil, err
	}
	return kvStates, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByKey(teamID int, key string) ([]*model.KVState, error) {
	var kvStates []*model.KVState
	if err := impl.db.Where("team_id = ? AND key = ?", teamID, key).Find(&kvStates).Error; err != nil {
		return nil, err
	}
	return kvStates, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*model.KVState, error) {
	var kvStates []*model.KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ?", teamID, apprefid, statetype, version).Find(&kvStates).Error; err != nil {
		return nil, err
	}
	return kvStates, nil
}

func (impl *KVStateStorage) RetrieveEditVersionByAppAndKey(teamID int, apprefid int, statetype int, key string) (*model.KVState, error) {
	var kvState *model.KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ? AND key = ?", teamID, apprefid, statetype, model.APP_EDIT_VERSION, key).First(&kvState).Error; err != nil {
		return nil, err
	}
	return kvState, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, version int) ([]*model.KVState, error) {
	var kvStates []*model.KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, version).Find(&kvStates).Error; err != nil {
		return nil, err
	}
	return kvStates, nil
}

func (impl *KVStateStorage) DeleteAllTypeKVStatesByApp(teamID int, apprefid int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, apprefid).Delete(&model.KVState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) DeleteAllKVStatesByAppVersionAndType(teamID int, apprefid int, version int, stateType int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ? AND state_type = ?", teamID, apprefid, version, stateType).Delete(&model.KVState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) DeleteAllTypeKVStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, targetVersion int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, targetVersion).Delete(&model.KVState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) DeleteAllTypeKVStatesByTeamIDAppIDAndVersionAndKey(teamID int, apprefid int, targetVersion int, key string) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ? AND key = ?", teamID, apprefid, targetVersion, key).Delete(&model.KVState{}).Error; err != nil {
		return err
	}
	return nil
}
