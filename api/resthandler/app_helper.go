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

package resthandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
	"github.com/illacloud/builder-backend/internal/auditlogger"
	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/action"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func (impl AppRestHandlerImpl) SnapshotTreeState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	// get edit version tree state from database
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(teamID, appID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	oldIDMap := map[int]int{}
	releaseIDMap := map[int]int{}

	// set version as mainline version
	for serial, _ := range treestates {
		oldIDMap[serial] = treestates[serial].ExportID()
		treestates[serial].AppendNewVersion(appMainLineVersion)
	}

	// put them to the database as duplicate, and record the old-new id map
	for i, treestate := range treestates {
		newID, errInCreateApp := impl.treestateRepository.Create(treestate)
		if err != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "create app failed: "+errInCreateApp.Error())
			return err
		}
		oldID := oldIDMap[i]
		releaseIDMap[oldID] = newID
	}

	// update children node ids
	for _, treestate := range treestates {
		treestate.RemapChildrenNodeRefIDs(releaseIDMap)
		treestate.SetParentNodeRefID(releaseIDMap[treestate.ParentNodeRefID])
		errInUpdateTreeState := impl.treestateRepository.Update(treestate)
		if errInUpdateTreeState != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_TREE_STATE, "update tree state failed: "+errInUpdateTreeState.Error())
			return errInUpdateTreeState
		}
	}

	return nil
}

func (impl AppRestHandlerImpl) SnapshotKVState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	// get edit version K-V state from database
	kvstates, errInRetrieveKVState := impl.kvstateRepository.RetrieveAllTypeKVStatesByApp(teamID, appID, repository.APP_EDIT_VERSION)
	if errInRetrieveKVState != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_KV_STATE, "get kv state failed: "+errInRetrieveKVState.Error())
		return errInRetrieveKVState
	}

	// set version as mainline version
	for serial, _ := range kvstates {
		kvstates[serial].AppendNewVersion(appMainLineVersion)
	}

	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		errInCreateKVState := impl.kvstateRepository.Create(kvstate)
		if errInCreateKVState != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_KV_STATE, "create kv state failed: "+errInCreateKVState.Error())
			return errInCreateKVState
		}
	}
	return nil
}

func (impl AppRestHandlerImpl) SnapshotSetState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	setstates, errInRetrieveSetState := impl.setstateRepository.RetrieveSetStatesByApp(teamID, appID, repository.SET_STATE_TYPE_DISPLAY_NAME, repository.APP_EDIT_VERSION)
	if errInRetrieveSetState != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_SET_STATE, "get set state failed: "+errInRetrieveSetState.Error())
		return errInRetrieveSetState
	}

	// update some fields
	for serial, _ := range setstates {
		setstates[serial].AppendNewVersion(mainlineVersion)
	}

	// and put them to the database as duplicate
	for _, setstate := range setstates {
		errInCreateSetState := impl.setstateRepository.Create(setstate)
		if errInCreateSetState != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_SET_STATE, "create set state failed: "+errInCreateSetState.Error())
			return errInCreateSetState
		}
	}
	return nil
}

func (impl AppRestHandlerImpl) SnapshotAction(c *gin.Context, teamID int, appID int, mainlineVersion int) error {
	// get edit version K-V state from database
	actions, errinRetrieveAction := impl.actionRepository.RetrieveActionsByAppVersion(teamID, appID, repository.APP_EDIT_VERSION)
	if errinRetrieveAction != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action failed: "+errinRetrieveAction.Error())
		return errinRetrieveAction
	}

	// set version as mainline version
	for serial, _ := range actions {
		actions[serial].AppendNewVersion(mainlineVersion)
	}

	// and put them to the database as duplicate
	for _, action := range actions {
		_, errInCreateAction := impl.actionRepository.Create(action)
		if errInCreateAction != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "create action failed: "+errInCreateAction.Error())
			return errInCreateAction
		}
	}
	return nil
}

func (impl AppRestHandlerImpl) TakeSnapshot(c *gin.Context, teamID int, appID int, mainlineVersion int, snapshotTriggerMode int) error {
	// retrieve app mainline version snapshot
	editVersionAppSnapshot, errInRetrieveSnapshot := impl.AppSnapshotRepository.RetrieveByTeamIDAppIDAndTargetVersion(teamID, appID, repository.APP_EDIT_VERSION)
	if errInRetrieveSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return errInRetrieveSnapshot
	}

	// set mainline version
	editVersionAppSnapshot.SetTargetVersion(mainlineVersion)

	// update old edit version snapshot
	errInUpdateSnapshot := impl.AppSnapshotRepository.UpdateWholeSnapshot(editVersionAppSnapshot)
	if errInUpdateSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT, "update snapshot failed: "+errInUpdateSnapshot.Error())
		return errInUpdateSnapshot
	}

	// create new edit version snapshot
	newAppSnapShot := repository.NewAppSnapshot(teamID, appID, repository.APP_EDIT_VERSION, snapshotTriggerMode)

	// storage new edit version snapshot
	_, errInCreateSnapshot := impl.AppSnapshotRepository.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "create snapshot failed: "+errInCreateSnapshot.Error())
		return errInCreateSnapshot
	}

	return nil
}

func (impl AppRestHandlerImpl) GetTargetVersionApp(c *gin.Context, teamID int, appID int, version int) (*repository.App, error) {
	// fetch app
	app, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app mega data error: "+errInRetrieveApp.Error())
		return errInRetrieveApp
	}

	// set for auto-version
	if version == repository.APP_AUTO_MAINLINE_VERSION {
		version = app.MainlineVersion
	}
	if version == repository.APP_AUTO_RELEASE_VERSION {
		version = app.ReleaseVersion
	}

	// form editor object field appForExport
	//
	// We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// get all modifier user ids from all apps
	allUserIDs := repository.ExtractAllEditorIDFromApps([]*repository.App{app})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return errInGetMultiUserInfo
	}

	appForExport := repository.NewAppForExport(app, usersLT)

	// form editor object field actions
	actions, errInRetrieveActions := impl.ActionRepository.RetrieveActionsByAppVersion(teamID, appID, version)
	if errInRetrieveActions != nil {
		actions = []*repository.Action{}
	}
	actionsForExport := make([]*repository.ActionForExport, 0)
	for _, action := range actions {
		actionsForExport = append(actionsForExport, repository.NewActionForExport(action))
	}

	// form editor object field components
	treeStateComponents, errInRetrieveTreeStateComponents := impl.TreeStateRepository.RetrieveTreeStatesByApp(teamID, appID, repository.TREE_STATE_TYPE_COMPONENTS, version)
	if errInRetrieveTreeStateComponents != nil {
		treeStateComponents = []*repository.TreeState{}
	}
	treeStateLT := repository.BuildTreeStateLookupTable(treeStateComponents)
	rootOfTreeState := repository.PickUpTreeStatesRootNode(treeStateComponents)
	componentTree, _ := repository.BuildComponentTree(rootOfTreeState, treeStateLT, nil)

	// form editor object field dependenciesState
	dependenciesState := map[string][]string{}
	dependenciesKVStates, errInRetrieveDependenciesKVStates := impl.KVStateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DEPENDENCIES, version)
	if errInRetrieveDependenciesKVStates != nil {
		dependenciesKVStates = []*repository.KVState{}
	}
	for _, dependency := range dependenciesKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dependency.Value), &revMsg)
		dependenciesState[dependency.Key] = revMsg // value convert to []string
	}

	// form editor object field dragShadowState
	dragShadowState := map[string]interface{}{}
	dragShadowKVStates, errInRetrieveDragShadowKVStates := impl.KVStateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DRAG_SHADOW, version)
	if errInRetrieveDragShadowKVStates != nil {
		dragShadowKVStates = []*repository.KVState{}
	}
	for _, dragShadow := range dragShadowKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dragShadow.Value), &revMsg)
		dragShadowState[dragShadow.Key] = revMsg // value convert to []string
	}

	// form editor object field dottedLineSquareState
	dottedLineSquareState := map[string]interface{}{}
	dottedLineSquareKVStates, errInRetrieveDottedLineSquareKVStates := impl.KVStateRepository.RetrieveKVStatesByApp(teamID, appID, repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE, version)
	if errInRetrieveDottedLineSquareKVStates != nil {
		dottedLineSquareKVStates = []*repository.KVState{}
	}
	for _, dottedLineSquare := range dottedLineSquareKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dottedLineSquare.Value), &revMsg)
		dottedLineSquareState[dottedLineSquare.Key] = revMsg // value convert to []string
	}

	// form editor object field displayNameState
	displayNameSetStates, errInRetrieveDisplayNameSetState := impl.SetStateRepository.RetrieveSetStatesByApp(teamID, appID, repository.SET_STATE_TYPE_DISPLAY_NAME, version)
	if errInRetrieveDisplayNameSetState != nil {
		displayNameSetStates = []*repository.SetState{}
	}
	displayNameState := make([]string, 0, len(displayNameSetStates))
	for _, displayName := range displayNameSetStates {
		displayNameState = append(displayNameState, displayName.Value)
	}

	// finally, make a brand new editor object
	editorForExport := repository.NewEditorForExport(appForExport, actionsForExport, componentTree, dependenciesState, dragShadowState, dottedLineSquareState, displayNameState)

	// feedback
	FeedbackOK(c, editorForExport)
	return app, nil
}
