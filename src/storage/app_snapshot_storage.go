package storage

import (
	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppSnapshotStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAppSnapshotStorage(logger *zap.SugaredLogger, db *gorm.DB) *AppSnapshotStorage {
	return &AppSnapshotStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *AppSnapshotStorage) Create(appSnapshot *model.AppSnapshot) (int, error) {
	if err := impl.db.Create(appSnapshot).Error; err != nil {
		return 0, err
	}
	return appSnapshot.ID, nil
}

func (impl *AppSnapshotStorage) RetrieveByID(id int) (*model.AppSnapshot, error) {
	var appSnapshot *model.AppSnapshot
	if err := impl.db.Where("id = ?", id).First(&appSnapshot).Error; err != nil {
		return nil, err
	}
	return appSnapshot, nil
}

func (impl *AppSnapshotStorage) RetrieveAll(teamID int) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Where("team_id = ?", teamID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}

func (impl *AppSnapshotStorage) RetrieveByTeamIDAndAppID(teamID int, appID int) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}

func (impl *AppSnapshotStorage) RetrieveCountByTeamIDAndAppID(teamID int, appID int) (int64, error) {
	var count int64
	if err := impl.db.Model(&AppSnapshot{}).Where("team_id = ? AND app_ref_id = ?", teamID, appID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (impl *AppSnapshotStorage) RetrieveEditVersion(teamID int, appID int) (*model.AppSnapshot, error) {
	var appSnapshot *model.AppSnapshot
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND target_version = ?", teamID, appID, APP_EDIT_VERSION).First(&appSnapshot).Error; err != nil {
		return nil, err
	}
	return appSnapshot, nil
}

func (impl *AppSnapshotStorage) RetrieveByTeamIDAppIDAndTargetVersion(teamID int, appID int, targetVersion int) (*model.AppSnapshot, error) {
	var appSnapshot *model.AppSnapshot
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND target_version = ?", teamID, appID, targetVersion).First(&appSnapshot).Error; err != nil {
		return nil, err
	}
	return appSnapshot, nil
}

func (impl *AppSnapshotStorage) RetrieveByTeamIDAppIDAndPage(teamID int, appID int, pagination *Pagination) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Scopes(paginate(impl.db, pagination)).Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}

func (impl *AppSnapshotStorage) UpdateWholeSnapshot(appSnapshot *model.AppSnapshot) error {
	if err := impl.db.Model(appSnapshot).Where("id = ?", appSnapshot.ID).UpdateColumns(appSnapshot).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionStorage) DeleteAllAppSnapshotByTeamIDAndAppID(teamID int, appID int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Delete(&model.AppSnapshot{}).Error; err != nil {
		return err
	}
	return nil
}
