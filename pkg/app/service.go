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

package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/idconvertor"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/internal/util/resourcelist"

	"go.uber.org/zap"
)

type AppService interface {
	CreateApp(app AppDto) (AppDto, error)
	IsPublicApp(teamID int, appID int) bool
	FetchAppByID(teamID int, appID int) (AppDto, error)
	GetAllApps(teamID int) ([]*AppDtoForExport, error)
	DuplicateApp(teamID int, appID, userID int, name string) (int, error)
	ReleaseApp(teamID int, appID int, userID int, public bool) (int, error)
	GetMegaData(teamID int, appID int, version int) (*EditorForExport, error)
	ReleaseTreeStateByApp(teamID int, appID int, mainlineVersion int) error
	ReleaseKVStateByApp(teamID int, appID int, mainlineVersion int) error
	ReleaseSetStateByApp(teamID int, appID int, mainlineVersion int) error
	ReleaseActionsByApp(teamID int, appID int, mainlineVersion int) error
}

type AppServiceImpl struct {
	logger              *zap.SugaredLogger
	appRepository       repository.AppRepository
	kvstateRepository   repository.KVStateRepository
	treestateRepository repository.TreeStateRepository
	setstateRepository  repository.SetStateRepository
	actionRepository    repository.ActionRepository
}

type AppActivity struct {
	Modifier   string    `json:"modifier"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type Editor struct {
	AppInfo               AppDto                 `json:"appInfo"`
	Actions               []Action               `json:"actions"`
	Components            *ComponentNode         `json:"components"`
	DependenciesState     map[string][]string    `json:"dependenciesState"`
	DragShadowState       map[string]interface{} `json:"dragShadowState"`
	DottedLineSquareState map[string]interface{} `json:"dottedLineSquareState"`
	DisplayNameState      []string               `json:"displayNameState"`
}

type EditorForExport struct {
	AppInfo               *AppDtoForExport       `json:"appInfo"`
	Actions               []*ActionForExport     `json:"actions"`
	Components            *ComponentNode         `json:"components"`
	DependenciesState     map[string][]string    `json:"dependenciesState"`
	DragShadowState       map[string]interface{} `json:"dragShadowState"`
	DottedLineSquareState map[string]interface{} `json:"dottedLineSquareState"`
	DisplayNameState      []string               `json:"displayNameState"`
}

func (resp *EditorForExport) ExportForFeedback() interface{} {
	return resp
}

func NewEditorForExport(editor *Editor) *EditorForExport {
	actionForExportSlice := make([]*ActionForExport, 0, len(editor.Actions))
	for _, action := range editor.Actions {
		actionForExportSlice = append(actionForExportSlice, NewActionForExport(&action))
	}
	return &EditorForExport{
		AppInfo:               NewAppDtoForExport(&editor.AppInfo),
		Actions:               actionForExportSlice,
		Components:            editor.Components,
		DependenciesState:     editor.DependenciesState,
		DragShadowState:       editor.DragShadowState,
		DottedLineSquareState: editor.DottedLineSquareState,
		DisplayNameState:      editor.DisplayNameState,
	}
}

type Action struct {
	ID          int                    `json:"actionId"`
	UID         uuid.UUID              `json:"uid"`
	TeamID      int                    `json:"teamID"`
	App         int                    `json:"-"`
	Version     int                    `json:"-"`
	Resource    int                    `json:"resourceId,omitempty"`
	DisplayName string                 `json:"displayName"`
	Type        string                 `json:"actionType"`
	Template    map[string]interface{} `json:"content"`
	Transformer map[string]interface{} `json:"transformer"`
	TriggerMode string                 `json:"triggerMode"`
	Config      string                 `json:"config"`
	CreatedAt   time.Time              `json:"createdAt,omitempty"`
	CreatedBy   int                    `json:"createdBy,omitempty"`
	UpdatedAt   time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy   int                    `json:"updatedBy,omitempty"`
}

func (action *Action) ExportConfig() *repository.ActionConfig {
	ac := &repository.ActionConfig{}
	json.Unmarshal([]byte(action.Config), ac)
	return ac
}

type ActionForExport struct {
	ID          string                   `json:"actionId"`
	UID         uuid.UUID                `json:"uid"`
	TeamID      string                   `json:"teamID"`
	App         string                   `json:"-"`
	Version     int                      `json:"-"`
	Resource    string                   `json:"resourceId,omitempty"`
	DisplayName string                   `json:"displayName"`
	Type        string                   `json:"actionType"`
	Template    map[string]interface{}   `json:"content"`
	Transformer map[string]interface{}   `json:"transformer"`
	TriggerMode string                   `json:"triggerMode"`
	Config      *repository.ActionConfig `json:"config"`
	CreatedAt   time.Time                `json:"createdAt,omitempty"`
	CreatedBy   string                   `json:"createdBy,omitempty"`
	UpdatedAt   time.Time                `json:"updatedAt,omitempty"`
	UpdatedBy   string                   `json:"updatedBy,omitempty"`
}

func NewActionForExport(action *Action) *ActionForExport {
	return &ActionForExport{
		ID:          idconvertor.ConvertIntToString(action.ID),
		UID:         action.UID,
		TeamID:      idconvertor.ConvertIntToString(action.TeamID),
		App:         idconvertor.ConvertIntToString(action.App),
		Version:     action.Version,
		Resource:    idconvertor.ConvertIntToString(action.Resource),
		DisplayName: action.DisplayName,
		Type:        action.Type,
		Template:    action.Template,
		Transformer: action.Transformer,
		TriggerMode: action.TriggerMode,
		Config:      action.ExportConfig(),
		CreatedAt:   action.CreatedAt,
		CreatedBy:   idconvertor.ConvertIntToString(action.CreatedBy),
		UpdatedAt:   action.UpdatedAt,
		UpdatedBy:   idconvertor.ConvertIntToString(action.UpdatedBy),
	}
}

func NewAppServiceImpl(logger *zap.SugaredLogger, appRepository repository.AppRepository,
	kvstateRepository repository.KVStateRepository,
	treestateRepository repository.TreeStateRepository, setstateRepository repository.SetStateRepository,
	actionRepository repository.ActionRepository) *AppServiceImpl {
	return &AppServiceImpl{
		logger:              logger,
		appRepository:       appRepository,
		kvstateRepository:   kvstateRepository,
		treestateRepository: treestateRepository,
		setstateRepository:  setstateRepository,
		actionRepository:    actionRepository,
	}
}

func (impl *AppServiceImpl) CreateApp(app AppDto) (AppDto, error) {
	// init
	app.ReleaseVersion = 0 // the draft version will always be 0, so the release version and mainline version are 0 by default when app init.
	app.MainlineVersion = 0
	app.CreatedAt = time.Now().UTC()
	app.UpdatedAt = time.Now().UTC()
	id, err := impl.appRepository.Create(&repository.App{
		UID:             app.UID,
		TeamID:          app.TeamID,
		Name:            app.Name,
		ReleaseVersion:  app.ReleaseVersion,
		MainlineVersion: app.MainlineVersion,
		Config:          app.Config.ExportToJSONString(),
		CreatedBy:       app.CreatedBy,
		CreatedAt:       app.CreatedAt,
		UpdatedBy:       app.UpdatedBy,
		UpdatedAt:       app.UpdatedAt,
	})
	if err != nil {
		return AppDto{}, err
	}
	// fetch user
	user, errInGetUserInfo := datacontrol.GetUserInfo(app.UpdatedBy)
	if errInGetUserInfo != nil {
		return AppDto{}, errInGetUserInfo
	}
	// fill data
	app.ID = id
	app.AppActivity.Modifier = user.Nickname
	app.AppActivity.ModifiedAt = app.UpdatedAt // TODO: find last modified time in another record with version 0
	return app, nil
}

func (impl *AppServiceImpl) IsPublicApp(teamID int, appID int) bool {
	app, err := impl.appRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if err != nil {
		return false
	}
	return app.IsPublic()
}

func (impl *AppServiceImpl) UpdateApp(app AppDto) (*AppDtoForExport, error) {
	app.UpdatedAt = time.Now().UTC()
	if err := impl.appRepository.Update(&repository.App{
		ID:              app.ID,
		UID:             app.UID,
		TeamID:          app.TeamID,
		Name:            app.Name,
		ReleaseVersion:  app.ReleaseVersion,
		MainlineVersion: app.MainlineVersion,
		Config:          app.Config.ExportToJSONString(),
		UpdatedBy:       app.UpdatedBy,
		UpdatedAt:       app.UpdatedAt,
	}); err != nil {
		return nil, err
	}

	// fetch user remote data
	user, errInGetUserInfo := datacontrol.GetUserInfo(app.UpdatedBy)
	if errInGetUserInfo != nil {
		return nil, errInGetUserInfo
	}

	// fill data
	app.AppActivity.Modifier = user.Nickname
	app.AppActivity.ModifiedAt = app.UpdatedAt // TODO: find last modified time in another record with version 0

	// feedback
	return NewAppDtoForExport(&app), nil
}

// call this method when action (over HTTP) and state (over websocket) changed
func (impl *AppServiceImpl) UpdateAppModifyTime(app *AppDto) error {
	app.UpdatedAt = time.Now().UTC()
	if err := impl.appRepository.UpdateUpdatedAt(&repository.App{
		ID:        app.ID,
		UpdatedBy: app.UpdatedBy,
		UpdatedAt: app.UpdatedAt,
	}); err != nil {
		return err
	}
	return nil
}

func (impl *AppServiceImpl) FetchAppByID(teamID int, appID int) (AppDto, error) {
	app, err := impl.appRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if err != nil {
		return AppDto{}, err
	}
	appDto := AppDto{
		ID:              app.ID,
		UID:             app.UID,
		TeamID:          app.TeamID,
		Name:            app.Name,
		ReleaseVersion:  app.ReleaseVersion,
		MainlineVersion: app.MainlineVersion,
		Config:          app.ExportConfig(),
		UpdatedBy:       app.UpdatedBy,
		UpdatedAt:       app.UpdatedAt,
	}
	return appDto, nil
}

func (impl *AppServiceImpl) DeleteApp(teamID int, appID int) error { // TODO: maybe need transaction
	_ = impl.treestateRepository.DeleteAllTypeTreeStatesByApp(teamID, appID)
	_ = impl.kvstateRepository.DeleteAllTypeKVStatesByApp(teamID, appID)
	_ = impl.actionRepository.DeleteActionsByApp(teamID, appID)
	_ = impl.setstateRepository.DeleteAllTypeSetStatesByApp(teamID, appID)
	return impl.appRepository.Delete(teamID, appID)
}

func (impl *AppServiceImpl) GetAllApps(teamID int) ([]*AppDtoForExport, error) {
	res, err := impl.appRepository.RetrieveAllByUpdatedTime(teamID)
	if err != nil {
		return nil, err
	}
	AppDtoForExportSlice := make([]*AppDtoForExport, 0, len(res))
	for _, value := range res {
		// fetch user remote data
		user, errInGetUserInfo := datacontrol.GetUserInfo(value.UpdatedBy)
		if errInGetUserInfo != nil {
			return nil, errInGetUserInfo
		}
		// fill data
		appDto := AppDto{
			ID:              value.ID,
			UID:             value.UID,
			TeamID:          value.TeamID,
			Name:            value.Name,
			ReleaseVersion:  value.ReleaseVersion,
			MainlineVersion: value.MainlineVersion,
			Config:          value.ExportConfig(),
			UpdatedAt:       value.UpdatedAt,
			UpdatedBy:       value.UpdatedBy,
			AppActivity: AppActivity{
				Modifier:   user.Nickname,
				ModifiedAt: value.UpdatedAt,
			},
		}
		AppDtoForExportSlice = append(AppDtoForExportSlice, NewAppDtoForExport(&appDto))

	}
	return AppDtoForExportSlice, nil
}

func (impl *AppServiceImpl) DuplicateApp(teamID int, appID int, userID int, name string) (int, error) {
	appA, err := impl.appRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if err != nil {
		return 0, err
	}
	appA.ReleaseVersion = 0 // the draft version will always be 0, so the release version and mainline version are 0 by default when app init.
	appA.MainlineVersion = 0
	appA.CreatedAt = time.Now().UTC()
	appA.UpdatedAt = time.Now().UTC()
	appA.CreatedBy = userID
	appA.UpdatedBy = userID
	newAppB := &repository.App{
		UID:             uuid.New(),
		TeamID:          teamID,
		Name:            name,
		ReleaseVersion:  appA.ReleaseVersion,
		MainlineVersion: appA.MainlineVersion,
		Config:          appA.Config,
		CreatedBy:       appA.CreatedBy,
		CreatedAt:       appA.CreatedAt,
		UpdatedBy:       appA.UpdatedBy,
		UpdatedAt:       appA.UpdatedAt,
	}
	newAppB.PushEditedBy(repository.NewAppEditedByUserID(userID))
	newAppB.SetPrivate()
	id, err := impl.appRepository.Create(newAppB)
	if err != nil {
		return 0, err
	}
	// fetch user remote data
	user, errInGetUserInfo := datacontrol.GetUserInfo(appA.UpdatedBy)
	if errInGetUserInfo != nil {
		return 0, errInGetUserInfo
	}
	// fill data
	appB := AppDto{
		ID:              id,
		Name:            name,
		ReleaseVersion:  appA.ReleaseVersion,
		MainlineVersion: appA.MainlineVersion,
		Config:          appA.ExportConfig(),
		CreatedBy:       appA.CreatedBy,
		CreatedAt:       appA.CreatedAt,
		UpdatedBy:       appA.UpdatedBy,
		UpdatedAt:       appA.UpdatedAt,
		AppActivity: AppActivity{
			Modifier:   user.Nickname,
			ModifiedAt: appA.UpdatedAt,
		},
	}
	_ = impl.copyAllTreeState(teamID, appID, appB.ID, userID)
	_ = impl.copyAllKVState(teamID, appID, appB.ID, userID)
	_ = impl.copyAllSetState(teamID, appID, appB.ID, userID)
	_ = impl.copyActions(teamID, appID, appB.ID, userID)

	// feedback
	return appB.ID, nil
}

func (impl *AppServiceImpl) copyAllTreeState(teamID, appA, appB, user int) error {
	// get edit version K-V state from database
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(teamID, appA, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// update some fields
	indexIDMap := map[int]int{}
	releaseIDMap := map[int]int{}
	for serial, _ := range treestates {
		indexIDMap[serial] = treestates[serial].ID
		treestates[serial].ID = 0
		treestates[serial].UID = uuid.New()
		treestates[serial].TeamID = teamID
		treestates[serial].AppRefID = appB
		treestates[serial].Version = repository.APP_EDIT_VERSION
		treestates[serial].CreatedBy = user
		treestates[serial].CreatedAt = time.Now().UTC()
		treestates[serial].UpdatedBy = user
		treestates[serial].UpdatedAt = time.Now().UTC()
	}
	// and put them to the database as duplicate
	for i, treestate := range treestates {
		id, err := impl.treestateRepository.Create(treestate)
		if err != nil {
			return err
		}
		oldID := indexIDMap[i]
		releaseIDMap[oldID] = id
	}

	for _, treestate := range treestates {
		treestate.ChildrenNodeRefIDs = convertLink(treestate.ChildrenNodeRefIDs, releaseIDMap)
		treestate.ParentNodeRefID = releaseIDMap[treestate.ParentNodeRefID]
		if err := impl.treestateRepository.Update(treestate); err != nil {
			return err
		}
	}

	return nil
}

func (impl *AppServiceImpl) copyAllKVState(teamID, appA, appB, user int) error {
	// get edit version K-V state from database
	kvstates, err := impl.kvstateRepository.RetrieveAllTypeKVStatesByApp(teamID, appA, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// update some fields
	for serial, _ := range kvstates {
		kvstates[serial].ID = 0
		kvstates[serial].UID = uuid.New()
		kvstates[serial].TeamID = teamID
		kvstates[serial].AppRefID = appB
		kvstates[serial].Version = repository.APP_EDIT_VERSION
		kvstates[serial].CreatedBy = user
		kvstates[serial].CreatedAt = time.Now().UTC()
		kvstates[serial].UpdatedBy = user
		kvstates[serial].UpdatedAt = time.Now().UTC()
	}
	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		if err := impl.kvstateRepository.Create(kvstate); err != nil {
			return err
		}
	}
	return nil
}

func (impl *AppServiceImpl) copyAllSetState(teamID, appA, appB, user int) error {
	setstates, err := impl.setstateRepository.RetrieveSetStatesByApp(teamID, appA, repository.SET_STATE_TYPE_DISPLAY_NAME, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// update some fields
	for serial, _ := range setstates {
		setstates[serial].ID = 0
		setstates[serial].UID = uuid.New()
		setstates[serial].TeamID = teamID
		setstates[serial].AppRefID = appB
		setstates[serial].Version = repository.APP_EDIT_VERSION
		setstates[serial].CreatedBy = user
		setstates[serial].CreatedAt = time.Now().UTC()
		setstates[serial].UpdatedBy = user
		setstates[serial].UpdatedAt = time.Now().UTC()
	}
	// and put them to the database as duplicate
	for _, setstate := range setstates {
		if err := impl.setstateRepository.Create(setstate); err != nil {
			return err
		}
	}
	return nil
}

func (impl *AppServiceImpl) copyActions(teamID, appA, appB, user int) error {
	// get edit version K-V state from database
	actions, err := impl.actionRepository.RetrieveActionsByAppVersion(teamID, appA, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// update some fields
	for serial, _ := range actions {
		actions[serial].ID = 0
		actions[serial].UID = uuid.New()
		actions[serial].TeamID = teamID
		actions[serial].App = appB
		actions[serial].Version = repository.APP_EDIT_VERSION
		actions[serial].CreatedBy = user
		actions[serial].CreatedAt = time.Now().UTC()
		actions[serial].UpdatedBy = user
		actions[serial].UpdatedAt = time.Now().UTC()
	}
	// and put them to the database as duplicate
	for _, action := range actions {
		if _, err := impl.actionRepository.Create(action); err != nil {
			return err
		}
	}
	return nil
}

func (impl *AppServiceImpl) ReleaseApp(teamID int, appID int, userID int, public bool) (int, error) {
	app, err := impl.appRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if err != nil {
		return -1, nil
	}

	// config app
	app.MainlineVersion += 1
	app.ReleaseVersion = app.MainlineVersion
	if public {
		app.SetPublic(userID)
	} else {
		app.SetPrivate(userID)
	}

	// exec release process
	_ = impl.ReleaseTreeStateByApp(teamID, appID, app.MainlineVersion)
	_ = impl.ReleaseKVStateByApp(teamID, appID, app.MainlineVersion)
	_ = impl.ReleaseSetStateByApp(teamID, appID, app.MainlineVersion)
	_ = impl.ReleaseActionsByApp(teamID, appID, app.MainlineVersion)
	if err := impl.appRepository.Update(app); err != nil {
		return -1, nil
	}

	return app.ReleaseVersion, nil
}

func (impl *AppServiceImpl) ReleaseTreeStateByApp(teamID int, appID int, mainlineVersion int) error {
	// get edit version K-V state from database
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(teamID, appID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	indexIDMap := map[int]int{}
	releaseIDMap := map[int]int{}
	// set version as mainline version
	for serial, _ := range treestates {
		indexIDMap[serial] = treestates[serial].ID
		treestates[serial].ID = 0
		treestates[serial].UID = uuid.New()
		treestates[serial].Version = mainlineVersion
	}
	// and put them to the database as duplicate
	for i, treestate := range treestates {
		id, err := impl.treestateRepository.Create(treestate)
		if err != nil {
			return err
		}
		oldID := indexIDMap[i]
		releaseIDMap[oldID] = id
	}
	for _, treestate := range treestates {
		treestate.ChildrenNodeRefIDs = convertLink(treestate.ChildrenNodeRefIDs, releaseIDMap)
		treestate.ParentNodeRefID = releaseIDMap[treestate.ParentNodeRefID]
		if err := impl.treestateRepository.Update(treestate); err != nil {
			return err
		}
	}

	return nil
}

func (impl *AppServiceImpl) ReleaseKVStateByApp(teamID int, appID int, mainlineVersion int) error {
	// get edit version K-V state from database
	kvstates, err := impl.kvstateRepository.RetrieveAllTypeKVStatesByApp(teamID, appID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// set version as mainline version
	for serial, _ := range kvstates {
		kvstates[serial].ID = 0
		kvstates[serial].UID = uuid.New()
		kvstates[serial].Version = mainlineVersion
	}
	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		if err := impl.kvstateRepository.Create(kvstate); err != nil {
			return err
		}
	}
	return nil
}

func (impl *AppServiceImpl) ReleaseSetStateByApp(teamID int, appID int, mainlineVersion int) error {
	setstates, err := impl.setstateRepository.RetrieveSetStatesByApp(teamID, appID, repository.SET_STATE_TYPE_DISPLAY_NAME, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// update some fields
	for serial, _ := range setstates {
		setstates[serial].ID = 0
		setstates[serial].UID = uuid.New()
		setstates[serial].Version = mainlineVersion
	}
	// and put them to the database as duplicate
	for _, setstate := range setstates {
		if err := impl.setstateRepository.Create(setstate); err != nil {
			return err
		}
	}
	return nil
}

func (impl *AppServiceImpl) ReleaseActionsByApp(teamID int, appID int, mainlineVersion int) error {
	// get edit version K-V state from database
	actions, err := impl.actionRepository.RetrieveActionsByAppVersion(teamID, appID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// set version as mainline version
	for serial, _ := range actions {
		actions[serial].ID = 0
		actions[serial].UID = uuid.New()
		actions[serial].Version = mainlineVersion
	}
	// and put them to the database as duplicate
	for _, action := range actions {
		if _, err := impl.actionRepository.Create(action); err != nil {
			return err
		}
	}
	return nil
}

func (impl *AppServiceImpl) GetMegaData(teamID, appID, version int) (*EditorForExport, error) {
	editor, err := impl.fetchEditor(teamID, appID, version)
	if err != nil {
		return nil, err
	}

	return NewEditorForExport(&editor), nil
}

func (impl *AppServiceImpl) fetchEditor(teamID int, appID int, version int) (Editor, error) {
	app, err := impl.appRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if err != nil {
		return Editor{}, err
	}
	if app.ID == 0 || version > app.MainlineVersion {
		return Editor{}, errors.New("content not found")
	}
	if version == repository.APP_AUTO_MAINLINE_VERSION {
		version = app.MainlineVersion
	}
	if version == repository.APP_AUTO_RELEASE_VERSION {
		version = app.ReleaseVersion
	}

	// fetch user remote data
	user, errInGetUserInfo := datacontrol.GetUserInfo(app.UpdatedBy)
	if errInGetUserInfo != nil {
		return Editor{}, errInGetUserInfo
	}
	// fill data
	res := Editor{}
	res.AppInfo = AppDto{
		ID:              app.ID,
		Name:            app.Name,
		ReleaseVersion:  app.ReleaseVersion,
		MainlineVersion: app.MainlineVersion,
		Config:          app.ExportConfig(),
		UpdatedAt:       app.UpdatedAt,
		UpdatedBy:       app.UpdatedBy,
		AppActivity: AppActivity{
			Modifier:   user.Nickname,
			ModifiedAt: app.UpdatedAt,
		},
	}
	res.Actions, _ = impl.formatActions(teamID, appID, version)
	res.Components, _ = impl.formatComponents(teamID, appID, version)
	res.DependenciesState, _ = impl.formatDependenciesState(teamID, appID, version)
	res.DragShadowState, _ = impl.formatDragShadowState(teamID, appID, version)
	res.DottedLineSquareState, _ = impl.formatDottedLineSquareState(teamID, appID, version)
	res.DisplayNameState, _ = impl.formatDisplayNameState(teamID, appID, version)

	return res, nil
}

func (impl *AppServiceImpl) formatActions(teamID, appID, version int) ([]Action, error) {
	res, err := impl.actionRepository.RetrieveActionsByAppVersion(teamID, appID, version)
	fmt.Printf("res dump: %v\n", res)
	if err != nil {
		return nil, err
	}

	resSlice := make([]Action, 0, len(res))
	for _, value := range res {
		resSlice = append(resSlice, Action{
			ID:          value.ID,
			UID:         value.UID,
			TeamID:      value.TeamID,
			Resource:    value.Resource,
			DisplayName: value.Name,
			Type:        resourcelist.GetResourceIDMappedType(value.Type),
			Transformer: value.Transformer,
			TriggerMode: value.TriggerMode,
			Template:    value.Template,
			Config:      value.Config,
			CreatedBy:   value.CreatedBy,
			CreatedAt:   value.CreatedAt,
			UpdatedBy:   value.UpdatedBy,
			UpdatedAt:   value.UpdatedAt,
		})
	}
	return resSlice, nil
}

func (impl *AppServiceImpl) formatComponents(teamID, appID, version int) (*ComponentNode, error) {
	res, err := impl.treestateRepository.RetrieveTreeStatesByApp(teamID, appID, repository.TREE_STATE_TYPE_COMPONENTS, version)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, errors.New("no component")
	}

	tempMap := map[int]*repository.TreeState{}
	root := &repository.TreeState{}
	for _, component := range res {
		if component.Name == repository.TREE_STATE_SUMMIT_NAME {
			root = component
		}
		tempMap[component.ID] = component
	}
	resNode, _ := buildComponentTree(root, tempMap, nil)

	return resNode, nil
}

func (impl *AppServiceImpl) formatDependenciesState(teamID, appID, version int) (map[string][]string, error) {
	res, err := impl.kvstateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DEPENDENCIES, version)
	if err != nil {
		return nil, err
	}

	resMap := map[string][]string{}

	if len(res) == 0 {
		return resMap, nil
	}

	for _, dependency := range res {
		var revMsg []string
		json.Unmarshal([]byte(dependency.Value), &revMsg)
		resMap[dependency.Key] = revMsg // value convert to []string
	}

	return resMap, nil
}

func (impl *AppServiceImpl) formatDragShadowState(teamID, appID, version int) (map[string]interface{}, error) {
	res, err := impl.kvstateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DRAG_SHADOW, version)
	if err != nil {
		return nil, err
	}

	resMap := map[string]interface{}{}

	if len(res) == 0 {
		return resMap, nil
	}

	for _, shadow := range res {
		var revMsg map[string]interface{}
		json.Unmarshal([]byte(shadow.Value), &revMsg)
		resMap[shadow.Key] = revMsg
	}

	return resMap, nil
}

func (impl *AppServiceImpl) formatDottedLineSquareState(teamID, appID, version int) (map[string]interface{}, error) {
	res, err := impl.kvstateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE, version)
	if err != nil {
		return nil, err
	}

	resMap := map[string]interface{}{}

	if len(res) == 0 {
		return resMap, nil
	}

	for _, line := range res {
		var revMsg map[string]interface{}
		json.Unmarshal([]byte(line.Value), &revMsg)
		resMap[line.Key] = line.Value
	}

	return resMap, nil
}

func (impl *AppServiceImpl) formatDisplayNameState(teamID, appID, version int) ([]string, error) {
	res, err := impl.setstateRepository.RetrieveSetStatesByApp(teamID, appID, repository.SET_STATE_TYPE_DISPLAY_NAME, version)
	if err != nil {
		return nil, err
	}

	resSlice := make([]string, 0, len(res))
	if len(res) == 0 {
		return resSlice, nil
	}

	for _, displayName := range res {
		resSlice = append(resSlice, displayName.Value)
	}

	return resSlice, nil
}

func convertLink(ref string, idMap map[int]int) string {
	// convert string to []int
	var oldIDs []int
	if err := json.Unmarshal([]byte(ref), &oldIDs); err != nil {
		return ""
	}
	// map old id to new id
	newIDs := make([]int, 0, len(oldIDs))
	for _, oldID := range oldIDs {
		newIDs = append(newIDs, idMap[oldID])
	}
	// convert []int to string
	idsjsonb, err := json.Marshal(newIDs)
	if err != nil {
		return ""
	}
	// return result
	return string(idsjsonb)
}
