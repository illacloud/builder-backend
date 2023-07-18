package repository

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppSnapshotRepository interface {
	Create(appSnapshot *AppSnapshot) (int, error)
	RetrieveByID(id int) (*AppSnapshot, error)
	RetrieveAll(teamID int) ([]*AppSnapshot, error)
	RetrieveByTeamIDAndAppID(teamID int, appID int) ([]*AppSnapshot, error)
	RetrieveByTeamIDAppIDAndTargetVersion(teamID int, appID int, targetVersion int) (*AppSnapshot, error)
	RetrieveByTeamIDAppIDAndPage(teamID int, appID int, pagination *Pagination) ([]*AppSnapshot, error)
	UpdateWholeSnapshot(appSnapshot *AppSnapshot) error
	RetrieveCountByTeamIDAndAppID(teamID int, appID int) (int64, error)
}

type AppSnapshotRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAppSnapshotRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *AppSnapshotRepositoryImpl {
	return &AppSnapshotRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *AppSnapshotRepositoryImpl) Create(appSnapshot *AppSnapshot) (int, error) {
	if err := impl.db.Create(appSnapshot).Error; err != nil {
		return 0, err
	}
	return appSnapshot.ID, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveByID(id int) (*AppSnapshot, error) {
	var appSnapshot *AppSnapshot
	if err := impl.db.Where("id = ?", id).First(&appSnapshot).Error; err != nil {
		return nil, err
	}
	return appSnapshot, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveAll(teamID int) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Where("team_id = ?", teamID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveByTeamIDAndAppID(teamID int, appID int) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveCountByTeamIDAndAppID(teamID int, appID int) (int64, error) {
	var count int64
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveEditVersion(teamID int, appID int) (*AppSnapshot, error) {
	var appSnapshot *AppSnapshot
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND target_version = ?", teamID, appID, APP_EDIT_VERSION).First(&appSnapshot).Error; err != nil {
		return nil, err
	}
	return appSnapshot, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveByTeamIDAppIDAndTargetVersion(teamID int, appID int, targetVersion int) (*AppSnapshot, error) {
	var appSnapshot *AppSnapshot
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND target_version = ?", teamID, appID, targetVersion).First(&appSnapshot).Error; err != nil {
		return nil, err
	}
	return appSnapshot, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveByTeamIDAppIDAndPage(teamID int, appID int, pagination *Pagination) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Scopes(paginate(impl.db, pagination)).Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}

func (impl *AppSnapshotRepositoryImpl) UpdateWholeSnapshot(appSnapshot *AppSnapshot) error {
	if err := impl.db.Model(appSnapshot).Where("id = ?", appSnapshot.ID).UpdateColumns(appSnapshot).Error; err != nil {
		return err
	}
	return nil
}
