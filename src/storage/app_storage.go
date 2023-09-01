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
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, appID).Delete(&model.App{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorage) Update(app *model.App) error {
	if err := impl.db.Model(app).UpdateColumns(model.App{
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

func (impl *AppStorage) RetrieveByID(appID int) (*model.App, error) {
	var app *model.App
	if err := impl.db.Where("id = ?", appID).First(&app).Error; err != nil {
		return nil, err
	}
	return app, nil
}

func (d *AppStorage) RetrieveByIDs(ids []int) ([]*model.App, error) {
	apps := []*model.App{}
	if err := d.db.Where("(id) IN ?", ids).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveByTeamID(teamID int) ([]*model.App, error) {
	var apps []*model.App
	if err := impl.db.Where("team_id = ?", teamID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveDeployedAppByTeamID(teamID int) ([]*model.App, error) {
	var apps []*model.App
	if err := impl.db.Where("team_id = ? and release_version <> ?", teamID, model.APP_EDIT_VERSION).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveDeployedAppByTeamIDOrderByUpdatedTime(teamID int) ([]*model.App, error) {
	var apps []*model.App
	if err := impl.db.Where("team_id = ? and release_version <> ?", teamID, model.APP_EDIT_VERSION).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveAppByTeamIDAndAppID(teamID int, appID int) (*model.App, error) {
	var app *model.App
	if err := impl.db.Where("id = ? AND team_id = ?", appID, teamID).Find(&app).Error; err != nil {
		return nil, err
	}
	return app, nil
}

func (impl *AppStorage) RetrieveByTeamIDOrderByUpdatedTime(teamID int) ([]*model.App, error) {
	var apps []*model.App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) UpdateUpdatedAt(app *model.App) error {
	if err := impl.db.Model(app).UpdateColumns(model.App{
		UpdatedBy: app.UpdatedBy,
		UpdatedAt: app.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppStorage) CountAPPByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&model.App{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *AppStorage) RetrieveAppLastModifiedTime(teamID int) (time.Time, error) {
	var app *model.App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&app).Error; err != nil {
		return time.Time{}, err
	}
	return app.ExportUpdatedAt(), nil
}

func (impl *AppStorage) CountByTeamIDAndKeywords(teamID int, keywords string) (int64, error) {
	var count int64
	fuzzyKeywords := "%" + keywords + "%"
	if err := impl.db.Model(&model.App{}).Where("team_id = ? AND (name ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (impl *AppStorage) CountByTeamID(teamID int) (int64, error) {
	var count int64
	if err := impl.db.Model(&model.App{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (impl *AppStorage) RetrieveByTeamIDAndSortByCreatedAtDescByPage(teamID int, pagination *Pagination) ([]*model.App, error) {
	var apps []*model.App
	if err := impl.db.Scopes(paginate(impl.db, pagination)).Where("team_id = ?", teamID).Order("created_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveByTeamIDAndSortByUpdatedAtDescByPage(teamID int, pagination *Pagination) ([]*model.App, error) {
	var apps []*model.App
	if err := impl.db.Scopes(paginate(impl.db, pagination)).Where("team_id = ?", teamID).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveByKeywordsAndSortByCreatedAtDesc(teamID int, keywords string) ([]*model.App, error) {
	var apps []*model.App
	fuzzyKeywords := "%" + keywords + "%"
	if err := impl.db.Where("team_id = ? AND (name ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords).Order("created_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveByKeywordsAndSortByUpdatedAtDesc(teamID int, keywords string) ([]*model.App, error) {
	var apps []*model.App
	fuzzyKeywords := "%" + keywords + "%"
	if err := impl.db.Where("team_id = ? AND (name ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveByKeywordsAndSortByCreatedAtDescByPage(teamID int, keywords string, pagination *Pagination) ([]*model.App, error) {
	var apps []*model.App
	fuzzyKeywords := "%" + keywords + "%"
	if err := impl.db.Scopes(paginate(impl.db, pagination)).Where("team_id = ? AND (name ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords).Order("created_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppStorage) RetrieveByKeywordsAndSortByUpdatedAtDescByPage(teamID int, keywords string, pagination *Pagination) ([]*model.App, error) {
	var apps []*model.App
	fuzzyKeywords := "%" + keywords + "%"
	if err := impl.db.Scopes(paginate(impl.db, pagination)).Where("team_id = ? AND (name ilike ? OR config->>'description' ilike ?)", teamID, fuzzyKeywords, fuzzyKeywords).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}
