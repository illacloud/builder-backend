package model

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type KVStateRepository interface {
	Create(kvstate *KVState) error
	Delete(teamID int, kvstateID int) error
	Update(kvstate *KVState) error
	RetrieveByID(teamID int, kvstateID int) (*KVState, error)
	RetrieveKVStatesByVersion(teamID int, versionID int) ([]*KVState, error)
	RetrieveKVStatesByKey(teamID int, key string) ([]*KVState, error)
	RetrieveKVStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*KVState, error)
	RetrieveEditVersionByAppAndKey(teamID int, apprefid int, statetype int, key string) (*KVState, error)
	RetrieveAllTypeKVStatesByApp(teamID int, apprefid int, version int) ([]*KVState, error)
	DeleteAllTypeKVStatesByApp(teamID int, apprefid int) error
	DeleteAllKVStatesByAppVersionAndType(teamID int, apprefid int, version int, stateType int) error
}

type KVStateRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewKVStateRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *KVStateRepositoryImpl {
	return &KVStateRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *KVStateRepositoryImpl) Create(kvstate *KVState) error {
	if err := impl.db.Create(kvstate).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateRepositoryImpl) Delete(teamID int, kvstateID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", kvstateID, teamID).Delete(&KVState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateRepositoryImpl) Update(kvstate *KVState) error {
	if err := impl.db.Model(&KVState{}).Where("id = ?", kvstate.ID).UpdateColumns(kvstate).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateRepositoryImpl) RetrieveByID(teamID int, kvstateID int) (*KVState, error) {
	var kvstate *KVState
	if err := impl.db.Where("id = ? AND team_id = ?", kvstateID, teamID).First(&kvstate).Error; err != nil {
		return &KVState{}, err
	}
	return kvstate, nil
}

func (impl *KVStateRepositoryImpl) RetrieveKVStatesByVersion(teamID int, version int) ([]*KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND version = ?", teamID, version).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
}

func (impl *KVStateRepositoryImpl) RetrieveKVStatesByKey(teamID int, key string) ([]*KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND key = ?", teamID, key).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
}

func (impl *KVStateRepositoryImpl) RetrieveKVStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ?", teamID, apprefid, statetype, version).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
}

func (impl *KVStateRepositoryImpl) RetrieveEditVersionByAppAndKey(teamID int, apprefid int, statetype int, key string) (*KVState, error) {
	var kvstate *KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ? AND key = ?", teamID, apprefid, statetype, APP_EDIT_VERSION, key).First(&kvstate).Error; err != nil {
		return nil, err
	}
	return kvstate, nil
}

func (impl *KVStateRepositoryImpl) RetrieveAllTypeKVStatesByApp(teamID int, apprefid int, version int) ([]*KVState, error) {
	var kvstates []*KVState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, version).Find(&kvstates).Error; err != nil {
		return nil, err
	}
	return kvstates, nil
}

func (impl *KVStateRepositoryImpl) DeleteAllTypeKVStatesByApp(teamID int, apprefid int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, apprefid).Delete(&KVState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *KVStateRepositoryImpl) DeleteAllKVStatesByAppVersionAndType(teamID int, apprefid int, version int, stateType int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ? AND state_type = ?", teamID, apprefid, version, stateType).Delete(&KVState{}).Error; err != nil {
		return err
	}
	return nil
}
