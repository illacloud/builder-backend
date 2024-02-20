// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/src/request"
)

const APP_EDIT_VERSION = 0           // the editable version app ID always be 0
const APP_AUTO_MAINLINE_VERSION = -1 // -1 for get latest mainline version automatically
const APP_AUTO_RELEASE_VERSION = -2  // -1 for get latest release version automatically

const APP_FIELD_NAME = "appName"
const APP_EDITED_BY_MAX_LENGTH = 4

type App struct {
	ID              int       `json:"id" gorm:"column:id;type:bigserial;primary_key;unique"`
	UID             uuid.UUID `json:"uid" gorm:"column:uid;type:uuid;not null"`
	TeamID          int       `json:"teamID" gorm:"column:team_id;type:bigserial"`
	Name            string    `json:"name" gorm:"column:name;type:varchar"`
	ReleaseVersion  int       `json:"releaseVersion" gorm:"column:release_version;type:bigserial"`
	MainlineVersion int       `json:"mainlineVersion" gorm:"column:mainline_version;type:bigserial"`
	Config          string    `json:"config" gorm:"column:config;type:jsonb"`
	CreatedAt       time.Time `json:"createdAt" gorm:"column:created_at;type:timestamp"`
	CreatedBy       int       `json:"createdBy" gorm:"column:created_by;type:bigserial"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"column:updated_at;type:timestamp"`
	UpdatedBy       int       `json:"updatedBy" gorm:"column:updated_by;type:bigserial"`
	EditedBy        string    `json:"editedBy" gorm:"column:edited_by;type:jsonb"`
}

func NewAppByCreateAppRequest(req *request.CreateAppRequest, teamID int, modifyUserID int) (*App, error) {
	appConfig := NewAppConfig()
	errInSetAppType := appConfig.SetAppTypeByString(req.ExportAppType())
	if errInSetAppType != nil {
		return nil, errInSetAppType
	}
	app := &App{
		TeamID:          teamID,
		Name:            req.ExportAppName(),
		ReleaseVersion:  APP_EDIT_VERSION,
		MainlineVersion: APP_EDIT_VERSION,
		Config:          appConfig.ExportToJSONString(),
		CreatedBy:       modifyUserID,
		UpdatedBy:       modifyUserID,
	}
	app.PushEditedBy(NewAppEditedByUserID(modifyUserID))
	app.InitUID()
	app.InitCreatedAt()
	app.InitUpdatedAt()
	return app, nil
}

func NewAppForDuplicate(targetApp *App, newAppName string, modifyUserID int) *App {
	appConfig := NewAppConfig()
	appConfig.SetAppType(targetApp.ExportAppType())
	newApp := &App{
		TeamID:          targetApp.ExportTeamID(),
		Name:            newAppName,
		ReleaseVersion:  APP_EDIT_VERSION,
		MainlineVersion: APP_EDIT_VERSION,
		Config:          appConfig.ExportToJSONString(),
		CreatedBy:       modifyUserID,
		UpdatedBy:       modifyUserID,
	}
	newApp.SetPrivate(modifyUserID)
	newApp.PushEditedBy(NewAppEditedByUserID(modifyUserID))
	newApp.InitUID()
	newApp.InitCreatedAt()
	newApp.InitUpdatedAt()
	return newApp
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

func (app *App) InitForFork(teamID int, modifyUserID int) {
	app.ID = 0
	app.UID = uuid.New()
	app.TeamID = teamID
	app.InitCreatedAt()
	app.Modify(modifyUserID)
	app.SetNotPublishedToMarketplace(modifyUserID)
}

func (app *App) BumpMainlineVersion() {
	app.MainlineVersion += 1
}

func (app *App) BumpMainlineVersionOverReleaseVersion() {
	app.MainlineVersion += 1
	for {
		if app.MainlineVersion > app.ReleaseVersion {
			break
		}
		app.MainlineVersion += 1
	}
}

func (app *App) SyncMainlineVersionWithTreeStateLatestVersion(treeStateLatestVersion int) {
	if app.MainlineVersion < treeStateLatestVersion {
		app.MainlineVersion = treeStateLatestVersion
	}
}

func (app *App) ReleaseMainlineVersion() {
	app.ReleaseVersion = app.MainlineVersion
}

func (app *App) Release() {
	app.BumpMainlineVersion()
	app.ReleaseMainlineVersion()
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

func (app *App) IsPublishedToMarketplace() bool {
	ac := app.ExportConfig()
	return ac.PublishedToMarketplace
}

func (app *App) IsPublishWithAIAgent() bool {
	ac := app.ExportConfig()
	return ac.PublishWithAIAgent
}

func (app *App) ExportAppType() int {
	ac := app.ExportConfig()
	return ac.ExportAppType()
}

func (app *App) SetNotPublishedToMarketplace(userID int) {
	appConfig := app.ExportConfig()
	appConfig.SetNotPublishedToMarketplace()
	app.Config = appConfig.ExportToJSONString()
	app.UpdatedBy = userID
	app.PushEditedBy(NewAppEditedByUserID(userID))
	app.InitUpdatedAt()
}

func (app *App) SetPublishedToMarketplace(publishedToMarketpalce bool, modifyUserID int) {
	appConfig := app.ExportConfig()
	appConfig.PublishedToMarketplace = publishedToMarketpalce
	// app.Public will sync to publishedToMarketpalce
	appConfig.Public = publishedToMarketpalce
	app.Config = appConfig.ExportToJSONString()
	app.UpdatedBy = modifyUserID
	app.PushEditedBy(NewAppEditedByUserID(modifyUserID))
	app.InitUpdatedAt()
}

func (app *App) SetID(appID int) {
	app.ID = appID
}

func (app *App) SetPublic(userID int) {
	appConfig := app.ExportConfig()
	appConfig.Public = true
	app.UpdatedBy = userID
	app.InitUpdatedAt()
	app.PushEditedBy(NewAppEditedByUserID(userID))
	app.Config = appConfig.ExportToJSONString()
}

func (app *App) SetPrivate(userID int) {
	appConfig := app.ExportConfig()
	appConfig.Public = false
	app.UpdatedBy = userID
	app.InitUpdatedAt()
	app.PushEditedBy(NewAppEditedByUserID(userID))
	app.Config = appConfig.ExportToJSONString()
}

func (app *App) Modify(userID int) {
	app.UpdatedBy = userID
	app.InitUpdatedAt()
	app.PushEditedBy(NewAppEditedByUserID(userID))
}

// WARRING! this is an view-level method, do not use this method to sync database changes, just for export data.
func (app *App) RewritePublicSettings(isPublic bool) {
	appConfig := app.ExportConfig()
	appConfig.Public = isPublic
	app.Config = appConfig.ExportToJSONString()
}

func (app *App) ExportID() int {
	return app.ID
}

func (app *App) ExportAppName() string {
	return app.Name
}

func (app *App) ExportTeamID() int {
	return app.TeamID
}

func (app *App) ExportCreatedBy() int {
	return app.CreatedBy
}

func (app *App) ExportUpdatedBy() int {
	return app.UpdatedBy
}

func (app *App) ExportMainlineVersion() int {
	return app.MainlineVersion
}

func (app *App) ExportReleaseVersion() int {
	return app.ReleaseVersion
}

func (app *App) ExportModifiedAllUserIDs() []int {
	ret := make([]int, 0)
	appEditedBys := make([]*AppEditedBy, 0)
	json.Unmarshal([]byte(app.EditedBy), &appEditedBys)
	if len(appEditedBys) == 0 {
		return ret
	}
	// pick up user ids
	for _, appEditedBy := range appEditedBys {
		ret = append(ret, appEditedBy.UserID)
	}
	return ret
}

func (app *App) ExportEditedBy() []*AppEditedBy {
	appEditedBys := make([]*AppEditedBy, 0)
	json.Unmarshal([]byte(app.EditedBy), &appEditedBys)
	return appEditedBys
}

func (app *App) ImportEditedBy(appEditedBys []*AppEditedBy) {
	payload, _ := json.Marshal(appEditedBys)
	app.EditedBy = string(payload)
}

func (app *App) PushEditedBy(currentEditedBy *AppEditedBy) {
	editedByList := app.ExportEditedBy()
	fmt.Printf("[DUMP] PushEditedBy.editedByList: %+v\n ", editedByList)
	fmt.Printf("[DUMP] PushEditedBy.currentEditedBy: %+v\n ", currentEditedBy)
	// remove exists
	serial := 0
	for _, editedBy := range editedByList {
		if editedBy.UserID != currentEditedBy.UserID {
			editedByList[serial] = editedBy
			serial++
		}
	}
	editedByList = editedByList[:serial]

	// insert
	editedByList = append([]*AppEditedBy{currentEditedBy}, editedByList...)
	fmt.Printf("[DUMP] PushEditedBy.insert.editedByList: %+v\n ", editedByList)

	// check length
	if len(editedByList) > APP_EDITED_BY_MAX_LENGTH {
		editedByList = editedByList[:len(editedByList)-1]
	}

	fmt.Printf("[DUMP] finalEditedByList: \n")
	for serial, editedBy := range editedByList {
		fmt.Printf("    [%d] editedByList: %+v\n ", serial, editedBy)
	}

	// ok, set it
	app.ImportEditedBy(editedByList)
}

func (app *App) UpdateAppByConfigAppRawRequest(rawReq map[string]interface{}) error {
	assertPass := true
	for key, value := range rawReq {
		switch key {
		case APP_FIELD_NAME:
			app.Name, assertPass = value.(string)
			if !assertPass {
				return errors.New("update app name failed due to assert failed.")
			}
		default:
		}
	}
	return nil
}

func ExtractAllEditorIDFromApps(apps []*App) []int {
	// extract all target user id from apps
	allUserIDsHashT := make(map[int]int, 0)
	allUserIDs := make([]int, 0)
	for _, app := range apps {
		ids := app.ExportModifiedAllUserIDs()
		for _, id := range ids {
			allUserIDsHashT[id] = id
		}
		updatedByID := app.ExportUpdatedBy()
		allUserIDsHashT[updatedByID] = updatedByID
	}
	for _, id := range allUserIDsHashT {
		allUserIDs = append(allUserIDs, id)
	}
	return allUserIDs
}
