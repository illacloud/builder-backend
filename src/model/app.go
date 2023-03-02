package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const APP_EDIT_VERSION = 0           // the editable version app ID always be 0, app always init by edit version in default.
const APP_AUTO_MAINLINE_VERSION = -1 // -1 for get mainline version automatically
const APP_AUTO_RELEASE_VERSION = -2  // -1 for get release version automatically

type App struct {
	ID              int       `json:"id" 				gorm:"column:id;type:bigserial;primary_key;unique"`
	UID             uuid.UUID `json:"uid"   		    gorm:"column:uid;type:uuid;not null"`
	TeamID          int       `json:"teamID" 		    gorm:"column:team_id;type:bigserial"`
	Name            string    `json:"name" 				gorm:"column:name;type:varchar"`
	ReleaseVersion  int       `json:"releaseVersion" 	gorm:"column:release_version;type:bigserial"`
	MainlineVersion int       `json:"mainlineVersion" 	gorm:"column:mainline_version;type:bigserial"`
	Config          string    `json:"config" 	        gorm:"column:config;type:jsonb"`
	CreatedAt       time.Time `json:"created_at" 		gorm:"column:created_at;type:timestamp"`
	CreatedBy       int       `json:"created_by" 		gorm:"column:created_by;type:bigserial"`
	UpdatedAt       time.Time `json:"updated_at" 		gorm:"column:updated_at;type:timestamp"`
	UpdatedBy       int       `json:"updated_by" 		gorm:"column:updated_by;type:bigserial"`
}

func NewAppWithCreateAppRequest(teamID int, userID int, req *CreateAppRequest) *App {
	appConfig := NewAppConfig()
	app := &App{
		TeamID:          teamID,
		Name:            req.ExportName(),
		ReleaseVersion:  APP_EDIT_VERSION,
		MainlineVersion: APP_EDIT_VERSION,
		Config:          appConfig.ExportToJSONString(),
		CreatedBy:       userID,
		UpdatedBy:       userID,
	}
	app.InitUID()
	app.InitCreatedAt()
	app.InitUpdatedAt()
	return app
}

func (app *App) UpdateAppConfig(appConfig *AppConfig, userID int) {
	app.Config = appConfig.ExportToJSONString()
	app.UpdatedBy = userID
	app.InitUpdatedAt()
}

func (app *App) InitUID() {
	app.UID = uuid.New()
}

func (app *App) InitCreatedAt() {
	app.CreatedAt = time.Now().UTC()
}

func (app *App) InitUpdatedAt() {
	app.UpdatedAt = time.Now().UTC()
}

func (app *App) ExportUpdatedAt() time.Time {
	return app.UpdatedAt
}

func (app *App) ExportConfig() *AppConfig {
	ac := &AppConfig{}
	json.Unmarshal([]byte(app.Config), ac)
	return ac
}

func (app *App) IsPublic() bool {
	ac := app.ExportConfig()
	return ac.Public
}

func (app *App) SetPublic(userID int) {
	ac := app.ExportConfig()
	ac.Public = true
	app.UpdatedBy = userID
	app.InitUpdatedAt()
}

func (app *App) SetPrivate(userID int) {
	ac := app.ExportConfig()
	ac.Public = false
	app.UpdatedBy = userID
	app.InitUpdatedAt()
}
