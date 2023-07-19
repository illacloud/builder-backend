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

	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func (impl AppRestHandlerImpl) SnapshotTreeState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	return impl.DuplicateTreeStateByVersion(c, teamID, appID, repository.APP_EDIT_VERSION, appMainLineVersion)
}

func (impl AppRestHandlerImpl) SnapshotKVState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	return impl.DuplicateKVStateByVersion(c, teamID, appID, repository.APP_EDIT_VERSION, appMainLineVersion)
}

func (impl AppRestHandlerImpl) SnapshotSetState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	return impl.DuplicateSetStateByVersion(c, teamID, appID, repository.APP_EDIT_VERSION, appMainLineVersion)
}

func (impl AppRestHandlerImpl) SnapshotAction(c *gin.Context, teamID int, appID int, mainlineVersion int) error {
	return impl.DuplicateActionByVersion(c, teamID, appID, repository.APP_EDIT_VERSION, mainlineVersion)
}

// recover edit version treestate to target version (coby target version data to edit version)
func (impl AppRestHandlerImpl) DuplicateTreeStateByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	// get from version tree state from database
	treestates, errInRetrieveTreeState := impl.TreeStateRepository.RetrieveAllTypeTreeStatesByApp(teamID, appID, fromVersion)
	if errInRetrieveTreeState != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get tree state failed: "+errInRetrieveTreeState.Error())
		return errInRetrieveTreeState
	}
	oldIDMap := map[int]int{}
	releaseIDMap := map[int]int{}

	// set version to target version
	for serial, _ := range treestates {
		oldIDMap[serial] = treestates[serial].ExportID()
		treestates[serial].AppendNewVersion(toVersion)
	}

	// put them to the database as duplicate, and record the old-new id map
	for i, treestate := range treestates {
		newID, errInCreateApp := impl.TreeStateRepository.Create(treestate)
		if errInCreateApp != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "create app failed: "+errInCreateApp.Error())
			return errInCreateApp
		}
		oldID := oldIDMap[i]
		releaseIDMap[oldID] = newID
	}

	// update children node ids
	for _, treestate := range treestates {
		treestate.RemapChildrenNodeRefIDs(releaseIDMap)
		treestate.SetParentNodeRefID(releaseIDMap[treestate.ParentNodeRefID])
		errInUpdateTreeState := impl.TreeStateRepository.Update(treestate)
		if errInUpdateTreeState != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_TREE_STATE, "update tree state failed: "+errInUpdateTreeState.Error())
			return errInUpdateTreeState
		}
	}

	return nil
}

func (impl AppRestHandlerImpl) DuplicateKVStateByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	kvstates, errInRetrieveKVState := impl.KVStateRepository.RetrieveAllTypeKVStatesByApp(teamID, appID, fromVersion)
	if errInRetrieveKVState != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get kv state failed: "+errInRetrieveKVState.Error())
		return errInRetrieveKVState
	}

	// set version as mainline version
	for serial, _ := range kvstates {
		kvstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		errInCreateKVState := impl.KVStateRepository.Create(kvstate)
		if errInCreateKVState != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_STATE, "create kv state failed: "+errInCreateKVState.Error())
			return errInCreateKVState
		}
	}
	return nil
}

func (impl AppRestHandlerImpl) DuplicateSetStateByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	setstates, errInRetrieveSetState := impl.SetStateRepository.RetrieveSetStatesByApp(teamID, appID, repository.SET_STATE_TYPE_DISPLAY_NAME, fromVersion)
	if errInRetrieveSetState != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get set state failed: "+errInRetrieveSetState.Error())
		return errInRetrieveSetState
	}

	// update some fields
	for serial, _ := range setstates {
		setstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, setstate := range setstates {
		errInCreateSetState := impl.SetStateRepository.Create(setstate)
		if errInCreateSetState != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_STATE, "create set state failed: "+errInCreateSetState.Error())
			return errInCreateSetState
		}
	}
	return nil
}

func (impl AppRestHandlerImpl) DuplicateActionByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	actions, errinRetrieveAction := impl.ActionRepository.RetrieveActionsByAppVersion(teamID, appID, fromVersion)
	if errinRetrieveAction != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action failed: "+errinRetrieveAction.Error())
		return errinRetrieveAction
	}

	// set version as mainline version
	for serial, _ := range actions {
		actions[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, action := range actions {
		_, errInCreateAction := impl.ActionRepository.Create(action)
		if errInCreateAction != nil {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "create action failed: "+errInCreateAction.Error())
			return errInCreateAction
		}
	}
	return nil
}

func (impl AppRestHandlerImpl) SaveAppSnapshot(c *gin.Context, teamID int, appID int, userID int, mainlineVersion int, snapshotTriggerMode int) (*repository.AppSnapshot, error) {
	return impl.SaveAppSnapshotByVersion(c, teamID, appID, userID, repository.APP_EDIT_VERSION, mainlineVersion, snapshotTriggerMode)
}

func (impl AppRestHandlerImpl) InitAppSnapshot(c *gin.Context, teamID int, appID int) (*repository.AppSnapshot, error) {
	// / create new edit version snapshot
	newAppSnapShot := repository.NewAppSnapshot(teamID, appID, repository.APP_EDIT_VERSION, repository.SNAPSHOT_TRIGGER_MODE_AUTO)

	// storage new edit version snapshot
	_, errInCreateSnapshot := impl.AppSnapshotRepository.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "create snapshot failed: "+errInCreateSnapshot.Error())
		return nil, errInCreateSnapshot
	}
	return newAppSnapShot, nil
}

// SaveAppSnapshotByVersion() method do following process:
// - get current version snapshot
// - set it to target version
// - save it
// - create new empty snapshot for current version
func (impl AppRestHandlerImpl) SaveAppSnapshotByVersion(c *gin.Context, teamID int, appID int, userID int, fromVersion int, toVersion int, snapshotTriggerMode int) (*repository.AppSnapshot, error) {
	// retrieve app mainline version snapshot
	editVersionAppSnapshot, errInRetrieveSnapshot := impl.AppSnapshotRepository.RetrieveByTeamIDAppIDAndTargetVersion(teamID, appID, fromVersion)
	if errInRetrieveSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return nil, errInRetrieveSnapshot
	}

	// set mainline version
	editVersionAppSnapshot.SetTargetVersion(toVersion)

	// add modify history
	modifyHistoryLog := repository.NewTakeAppSnapshotModifyHistory(userID)
	editVersionAppSnapshot.PushModifyHistory(modifyHistoryLog)

	// update old edit version snapshot
	errInUpdateSnapshot := impl.AppSnapshotRepository.UpdateWholeSnapshot(editVersionAppSnapshot)
	if errInUpdateSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT, "update snapshot failed: "+errInUpdateSnapshot.Error())
		return nil, errInUpdateSnapshot
	}

	// create new edit version snapshot
	newAppSnapShot := repository.NewAppSnapshot(teamID, appID, fromVersion, snapshotTriggerMode)

	// storage new edit version snapshot
	_, errInCreateSnapshot := impl.AppSnapshotRepository.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "create snapshot failed: "+errInCreateSnapshot.Error())
		return nil, errInCreateSnapshot
	}

	return newAppSnapShot, nil
}

func (impl AppRestHandlerImpl) GetTargetVersionApp(c *gin.Context, teamID int, appID int, version int) (*repository.App, error) {
	// fetch app
	app, errInRetrieveApp := impl.AppRepository.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app mega data error: "+errInRetrieveApp.Error())
		return nil, errInRetrieveApp
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
		return nil, errInGetMultiUserInfo
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
