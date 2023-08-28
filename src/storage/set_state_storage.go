package storage

import (
	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SetStateStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewSetStateStorage(logger *zap.SugaredLogger, db *gorm.DB) *SetStateStorage {
	return &SetStateStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *SetStateStorage) Create(setState *model.SetState) error {
	if err := impl.db.Create(setState).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateStorage) Delete(teamID int, setStateID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", setStateID, teamID).Delete(&model.SetState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateStorage) DeleteByValue(setState *model.SetState) error {
	if err := impl.db.Where("team_id = ? AND value = ?", setState.TeamID, setState.Value).Delete(&model.SetState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateStorage) Update(setState *model.SetState) error {
	if err := impl.db.Model(setState).Where("id = ?", setState.ID).UpdateColumns(setState).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateStorage) UpdateByValue(beforeSetState *model.SetState, afterSetState *model.SetState) error {
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

func (impl *SetStateStorage) RetrieveByID(teamID int, setStateID int) (*model.SetState, error) {
	var setState *model.SetState
	if err := impl.db.Where("team_id = ? AND value = ?", teamID, setState.Value).First(&setState).Error; err != nil {
		return &model.SetState{}, err
	}
	return setState, nil
}

func (impl *SetStateStorage) RetrieveSetStatesByVersion(teamID int, version int) ([]*model.SetState, error) {
	var setStates []*model.SetState
	if err := impl.db.Where("team_id = ? AND version = ?", teamID, version).Find(&setStates).Error; err != nil {
		return nil, err
	}
	return setStates, nil
}

func (impl *SetStateStorage) RetrieveByValue(setState *model.SetState) (*model.SetState, error) {
	var ret *model.SetState
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

func (impl *SetStateStorage) RetrieveSetStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, statetype int, version int) ([]*model.SetState, error) {
	var setStates []*model.SetState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ?", teamID, apprefid, statetype, version).Find(&setStates).Error; err != nil {
		return nil, err
	}
	return setStates, nil
}

func (impl *SetStateStorage) DeleteAllTypeSetStatesByApp(teamID int, apprefid int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, apprefid).Delete(&model.SetState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateStorage) DeleteAllTypeSetStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, targetVersion int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, targetVersion).Delete(&model.SetState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *SetStateStorage) DeleteAllTypeSetStatesByTeamIDAppIDAndVersionAndValue(teamID int, apprefid int, targetVersion int, value string) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ? AND value = ?", teamID, apprefid, targetVersion, value).Delete(&model.SetState{}).Error; err != nil {
		return err
	}
	return nil
}
