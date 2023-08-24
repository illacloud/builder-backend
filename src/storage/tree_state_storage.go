package storage

import (
	"fmt"

	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TreeStateStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewTreeStateStorage(logger *zap.SugaredLogger, db *gorm.DB) *TreeStateStorage {
	return &TreeStateStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *TreeStateStorage) Create(treestate *model.TreeState) (int, error) {
	fmt.Printf("Createing tree_state: uid: %v, team_id: %v, app_id: %v, name: %v. \n", treestate.UID, treestate.TeamID, treestate.AppRefID, treestate.Name)
	if err := impl.db.Create(treestate).Error; err != nil {
		return 0, err
	}
	return treestate.ID, nil
}

func (impl *TreeStateStorage) Delete(teamID int, treestateID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", treestateID, teamID).Delete(&model.TreeState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) Update(treestate *model.TreeState) error {
	if err := impl.db.Model(treestate).UpdateColumns(TreeState{
		ID:                 treestate.ID,
		StateType:          treestate.StateType,
		ParentNodeRefID:    treestate.ParentNodeRefID,
		ChildrenNodeRefIDs: treestate.ChildrenNodeRefIDs,
		AppRefID:           treestate.AppRefID,
		Version:            treestate.Version,
		Name:               treestate.Name,
		Content:            treestate.Content,
		UpdatedAt:          treestate.UpdatedAt,
		UpdatedBy:          treestate.UpdatedBy,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) RetrieveByID(teamID int, treestateID int) (*model.TreeState, error) {
	treestate := &model.TreeState{}
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, treestateID).First(&treestate).Error; err != nil {
		return &model.TreeState{}, err
	}
	return treestate, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByVersion(teamID int, version int) ([]*model.TreeState, error) {
	var treestates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND version = ?", teamID, version).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesLatestVersion(teamID int, appID int) (int, error) {
	var treestates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Order("version desc").Limit(1).Find(&treestates).Error; err != nil {
		return 0, err
	}
	if len(treestates) == 0 {
		return 0, nil
	}
	return treestates[0].Version, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByName(teamID int, name string) ([]*model.TreeState, error) {
	var treestates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND name = ?", teamID, name).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*model.TreeState, error) {
	var treestates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ?", teamID, apprefid, statetype, version).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateStorage) RetrieveEditVersionByAppAndName(teamID int, apprefid int, statetype int, name string) (*model.TreeState, error) {
	var treestate *model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ? AND name = ?", teamID, apprefid, statetype, APP_EDIT_VERSION, name).First(&treestate).Error; err != nil {
		return nil, err
	}
	return treestate, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, version int) ([]*model.TreeState, error) {
	var treestates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, version).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateStorage) DeleteAllTypeTreeStatesByApp(teamID int, apprefid int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, apprefid).Delete(&model.TreeState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) DeleteAllTypeTreeStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, targetVersion int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, targetVersion).Delete(&model.TreeState{}).Error; err != nil {
		return err
	}
	return nil
}
