package storage

import (
	"time"

	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAppStorage(logger *zap.SugaredLogger, db *gorm.DB) *AppStorage {
	return &AppStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *AppStorage) Create(app *model.App) (int, error) {
	if err := impl.db.Create(app).Error; err != nil {
		return 0, err
	}
	return app.ID, nil
}

func (impl *AppStorage) Delete(teamID int, appID int) error {
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, appID).Delete(&App{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorage) Update(app *model.App) error {
	if err := impl.db.Model(app).UpdateColumns(App{
		Name:            app.Name,
		ReleaseVersion:  app.ReleaseVersion,
		MainlineVersion: app.MainlineVersion,
		Config:          app.Config,
		UpdatedBy:       app.UpdatedBy,
		UpdatedAt:       app.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorage) UpdateWholeApp(app *model.App) error {
	if err := impl.db.Model(app).Where("id = ?", app.ID).UpdateColumns(app).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorage) RetrieveAll(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveAppByIDAndTeamID(appID int, teamID int) (*App, error) {
	var app *App
	if err := impl.db.Where("id = ? AND team_id = ?", appID, teamID).Find(&app).Error; err != nil {
		return nil, err
	}
	return app, nil
}

func (impl *AppStorage) RetrieveAllByUpdatedTime(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) UpdateUpdatedAt(app *model.App) error {
	if err := impl.db.Model(app).UpdateColumns(App{
		UpdatedBy: app.UpdatedBy,
		UpdatedAt: app.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorage) CountAPPByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&App{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *AppStorage) RetrieveAppLastModifiedTime(teamID int) (time.Time, error) {
	var app *App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&app).Error; err != nil {
		return time.Time{}, err
	}
	return app.ExportUpdatedAt(), nil
}
