package model

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppStorage interface {
	Create(app *App) (int, error)
	Delete(teamID int, appID int) error
	Update(app *App) error
	RetrieveAll(teamID int) ([]*App, error)
	RetrieveAppByIDAndTeamID(appID int, teamID int) (*App, error)
	RetrieveAllByUpdatedTime(teamID int) ([]*App, error)
	CountAPPByTeamID(teamID int) (int, error)
	RetrieveAppLastModifiedTime(teamID int) (time.Time, error)
}

type AppStorageImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAppStorageImpl(logger *zap.SugaredLogger, db *gorm.DB) *AppStorageImpl {
	return &AppStorageImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *AppStorageImpl) Create(app *App) error {
	if err := impl.db.Create(app).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorageImpl) Delete(teamID int, appID int) error {
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, appID).Delete(&App{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorageImpl) Update(app *App) error {
	if err := impl.db.Model(&App{}).Where("id = ?", app.ID).UpdateColumns(app).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorageImpl) RetrieveAll(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorageImpl) RetrieveAppByIDAndTeamID(appID int, teamID int) (*App, error) {
	var app *App
	if err := impl.db.Where("id = ? AND team_id = ?", appID, teamID).Find(&app).Error; err != nil {
		return nil, err
	}
	return app, nil
}

func (impl *AppStorageImpl) RetrieveAllByUpdatedTime(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorageImpl) CountAPPByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&App{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *AppStorageImpl) RetrieveAppLastModifiedTime(teamID int) (time.Time, error) {
	var app *App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&app).Error; err != nil {
		return time.Time{}, err
	}
	return app.ExportUpdatedAt(), nil
}
