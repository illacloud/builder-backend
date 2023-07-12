package repository

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppSnapshotRepository interface {
	Create(appSnapshot *AppSnapshot) (int, error)
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
	return app.ID, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveAll(teamID int) ([]*App, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Where("team_id = ?", teamID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveByTeamIDAndAppID(teamID int, appID int) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}

func (impl *AppSnapshotRepositoryImpl) RetrieveByTeamIDAppIDAndPage(teamID int, appID int, pagination *Pagination) ([]*AppSnapshot, error) {
	var appSnapshots []*AppSnapshot
	if err := impl.db.Scope(paginate(impl.db, pagination)).Where("team_id = ? AND app_ref_id = ?", teamID, appID).Find(&appSnapshots).Error; err != nil {
		return nil, err
	}
	return appSnapshots, nil
}
