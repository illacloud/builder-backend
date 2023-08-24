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

func (impl *KVStateStorage) Create(kvstate *model.KVState) error {
	if err := impl.db.Create(kvstate).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) Delete(teamID int, kvstateID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", kvstateID, teamID).Delete(&model.KVState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) Update(kvstate *model.KVState) error {
	if err := impl.db.Model(kvstate).UpdateColumns(model.KVState{
		ID:        kvstate.ID,
		StateType: kvstate.StateType,
		AppRefID:  kvstate.AppRefID,
		Version:   kvstate.Version,
		Key:       kvstate.Key,
		Value:     kvstate.Value,
		UpdatedAt: kvstate.UpdatedAt,
		UpdatedBy: kvstate.UpdatedBy,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateStorage) RetrieveByID(teamID int, kvstateID int) (*model.KVState, error) {
	var kvstate *KVState
	if err := impl.db.Where("id = ? AND team_id = ?", kvstateID, teamID).First(&kvstate).Error; err != nil {
		return &KVState{}, err
	}
	return kvstate, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByVersion(teamID int, version int) ([]*model.KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND version = ?", teamID, version).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByKey(teamID int, key string) ([]*model.KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND key = ?", teamID, key).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*model.KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ?", teamID, apprefid, statetype, version).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
}

func (impl *KVStateStorage) RetrieveEditVersionByAppAndKey(teamID int, apprefid int, statetype int, key string) (*model.KVState, error) {
	var kvstate *KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ? AND key = ?", teamID, apprefid, statetype, APP_EDIT_VERSION, key).First(&kvstate).Error; err != nil {
		return nil, err
	}
	return kvstate, nil
}

func (impl *KVStateStorage) RetrieveKVStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, version int) ([]*model.KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, version).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
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
