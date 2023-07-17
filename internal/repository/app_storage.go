package repository

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppRepository interface {
	Create(app *App) (int, error)
	Delete(teamID int, appID int) error
	Update(app *App) error
	UpdateWholeApp(app *App) error
	UpdateUpdatedAt(app *App) error
	RetrieveAll(teamID int) ([]*App, error)
	RetrieveAppByIDAndTeamID(appID int, teamID int) (*App, error)
	RetrieveAllByUpdatedTime(teamID int) ([]*App, error)
	CountAPPByTeamID(teamID int) (int, error)
	RetrieveAppLastModifiedTime(teamID int) (time.Time, error)
}

type AppRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAppRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *AppRepositoryImpl {
	return &AppRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *AppRepositoryImpl) Create(app *App) (int, error) {
	if err := impl.db.Create(app).Error; err != nil {
		return 0, err
	}
	return app.ID, nil
}

func (impl *AppRepositoryImpl) Delete(teamID int, appID int) error {
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, appID).Delete(&App{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) Update(app *App) error {
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

func (impl *AppRepositoryImpl) UpdateWholeApp(app *App) error {
	if err := impl.db.Model(app).Where("id = ?", app.ID).UpdateColumns(app).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) RetrieveAll(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppRepositoryImpl) RetrieveAppByIDAndTeamID(appID int, teamID int) (*App, error) {
	var app *App
	if err := impl.db.Where("id = ? AND team_id = ?", appID, teamID).Find(&app).Error; err != nil {
		return nil, err
	}
	return app, nil
}

func (impl *AppRepositoryImpl) RetrieveAllByUpdatedTime(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppRepositoryImpl) UpdateUpdatedAt(app *App) error {
	if err := impl.db.Model(app).UpdateColumns(App{
		UpdatedBy: app.UpdatedBy,
		UpdatedAt: app.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) CountAPPByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&App{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *AppRepositoryImpl) RetrieveAppLastModifiedTime(teamID int) (time.Time, error) {
	var app *App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&app).Error; err != nil {
		return time.Time{}, err
	}
	return app.ExportUpdatedAt(), nil
}
