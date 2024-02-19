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

package controller

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/datacontrol"
	"github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"

	"github.com/gin-gonic/gin"
)

// recover edit version treeState to target version (copy target version data to edit version)
func (controller *Controller) DuplicateTreeStateByVersion(c *gin.Context, fromTeamID int, toTeamID int, fromAppID int, toAppID int, fromVersion int, toVersion int, modifierID int) error {
	// get target version tree state from database
	treeStates, errinRetrieveTreeStates := controller.Storage.TreeStateStorage.RetrieveTreeStatesByTeamIDAppIDAndVersion(fromTeamID, fromAppID, fromVersion)
	if errinRetrieveTreeStates != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get tree state failed: "+errinRetrieveTreeStates.Error())
		return errinRetrieveTreeStates
	}
	indexIDMap := map[int]int{}
	idConvertMap := map[int]int{}

	// set fork info
	for serial, _ := range treeStates {
		indexIDMap[serial] = treeStates[serial].ExportID()
		treeStates[serial].InitForFork(toTeamID, toAppID, toVersion, modifierID)
	}

	// put them to the database as duplicate, and record the old-new id map
	for i, treeState := range treeStates {
		treeStateID, errInCreateApp := controller.Storage.TreeStateStorage.Create(treeState)
		if errInCreateApp != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "create app failed: "+errInCreateApp.Error())
			return errInCreateApp
		}
		oldID := indexIDMap[i]
		idConvertMap[oldID] = treeStateID
	}

	// update tree states parent & children relation
	for _, treeState := range treeStates {
		treeState.ResetChildrenNodeRefIDsByMap(idConvertMap)
		treeState.ResetParentNodeRefIDByMap(idConvertMap)
		errInUpdateTreeState := controller.Storage.TreeStateStorage.Update(treeState)
		if errInUpdateTreeState != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_TREE_STATE, "update tree state failed: "+errInUpdateTreeState.Error())
			return errInUpdateTreeState
		}
	}

	return nil
}

func (controller *Controller) DuplicateKVStateByVersion(c *gin.Context, fromTeamID int, toTeamID int, fromAppID int, toAppID int, fromVersion int, toVersion int, modifierID int) error {
	// get target version K-V state from database
	kvStates, errInRetrieveKVStates := controller.Storage.KVStateStorage.RetrieveKVStatesByTeamIDAppIDAndVersion(fromTeamID, fromAppID, fromVersion)
	if errInRetrieveKVStates != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get kv state failed: "+errInRetrieveKVStates.Error())
		return errInRetrieveKVStates
	}

	// set fork info
	for serial, _ := range kvStates {
		kvStates[serial].InitForFork(toTeamID, toAppID, toVersion, modifierID)
	}

	// and put them to the database as duplicate
	for _, kvState := range kvStates {
		errInCreateKVState := controller.Storage.KVStateStorage.Create(kvState)
		if errInCreateKVState != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_STATE, "create kv state failed: "+errInCreateKVState.Error())
			return errInCreateKVState
		}
	}
	return nil
}

func (controller *Controller) DuplicateSetStateByVersion(c *gin.Context, fromTeamID int, toTeamID int, fromAppID int, toAppID int, fromVersion int, toVersion int, modifierID int) error {
	// get target version set state from database
	setStates, errInRetrieveSetStates := controller.Storage.SetStateStorage.RetrieveSetStatesByTeamIDAppIDAndVersion(fromTeamID, fromAppID, model.SET_STATE_TYPE_DISPLAY_NAME, fromVersion)
	if errInRetrieveSetStates != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get set state failed: "+errInRetrieveSetStates.Error())
		return errInRetrieveSetStates
	}

	// set fork info
	for serial, _ := range setStates {
		setStates[serial].InitForFork(toTeamID, toAppID, toVersion, modifierID)
	}

	// and put them to the database as duplicate
	for _, setState := range setStates {
		errInCreateSetState := controller.Storage.SetStateStorage.Create(setState)
		if errInCreateSetState != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_STATE, "create set state failed: "+errInCreateSetState.Error())
			return errInCreateSetState
		}
	}

	return nil
}

func (controller *Controller) DuplicateActionByVersion(c *gin.Context, fromTeamID int, toTeamID int, fromAppID int, toAppID int, fromVersion int, toVersion int, makeItPublic bool, modifierID int, isForkApp bool) error {
	// get target version action from database
	actions, errinRetrieveAction := controller.Storage.ActionStorage.RetrieveActionsByTeamIDAppIDAndVersion(fromTeamID, fromAppID, fromVersion)
	if errinRetrieveAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action failed: "+errinRetrieveAction.Error())
		return errinRetrieveAction
	}

	// set fork info
	for serial, _ := range actions {
		actions[serial].InitForFork(toTeamID, toAppID, toVersion, modifierID)
		if makeItPublic {
			actions[serial].SetPublic(modifierID)
		} else {
			actions[serial].SetPrivate(modifierID)
		}
	}

	// and put them to the database as duplicate
	resourceManagerSDK, errInNewResourceManagerSDK := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
	if errInNewResourceManagerSDK != nil {
		return errInNewResourceManagerSDK
	}
	for _, action := range actions {
		// check if action is ai-agent, and if ai-agent is public, and we are forking app from marketplace (not publish app to marketplace) fork it automatically
		if action.Type == resourcelist.TYPE_AI_AGENT_ID && isForkApp {
			fmt.Printf("[DUMP] DuplicateActionByVersion: hit AI_AGENT action\n")
			// call resource manager for for ai-agent
			forkedAIAgent, errInForkAiAgent := resourceManagerSDK.ForkMarketplaceAIAgent(action.ExportResourceID(), toTeamID, modifierID)
			fmt.Printf("[DUMP] DuplicateActionByVersion() forkedAIAgent: %+v\n", forkedAIAgent)
			fmt.Printf("[DUMP] DuplicateActionByVersion() errInForkAiAgent: %+v\n", errInForkAiAgent)
			if errInForkAiAgent == nil {
				action.SetResourceIDByAiAgent(forkedAIAgent)
			}
		}
		fmt.Printf("[DUMP] DuplicateActionByVersion() action: %+v\n", action)

		// create action
		_, errInCreateAction := controller.Storage.ActionStorage.Create(action)
		if errInCreateAction != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "create action failed: "+errInCreateAction.Error())
			return errInCreateAction
		}
	}
	return nil
}

func (controller *Controller) SaveAppSnapshot(c *gin.Context, teamID int, appID int, userID int, mainlineVersion int, snapshotTriggerMode int) (*model.AppSnapshot, error) {
	return controller.SaveAppSnapshotByVersion(c, teamID, appID, userID, model.APP_EDIT_VERSION, mainlineVersion, snapshotTriggerMode)
}

func (controller *Controller) InitAppSnapshot(c *gin.Context, teamID int, appID int) (*model.AppSnapshot, error) {
	// / create new edit version snapshot
	newAppSnapShot := model.NewAppSnapshot(teamID, appID, model.APP_EDIT_VERSION, model.SNAPSHOT_TRIGGER_MODE_AUTO)

	// storage new edit version snapshot
	_, errInCreateSnapshot := controller.Storage.AppSnapshotStorage.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "create snapshot failed: "+errInCreateSnapshot.Error())
		return nil, errInCreateSnapshot
	}
	return newAppSnapShot, nil
}

// SaveAppSnapshotByVersion() method do following process:
// - get current version snapshot
// - set it to target version
// - save it
// - create new empty snapshot for current version
func (controller *Controller) SaveAppSnapshotByVersion(c *gin.Context, teamID int, appID int, userID int, fromVersion int, toVersion int, snapshotTriggerMode int) (*model.AppSnapshot, error) {
	// retrieve app mainline version snapshot
	editVersionAppSnapshot, errInRetrieveSnapshot := controller.Storage.AppSnapshotStorage.RetrieveByTeamIDAppIDAndTargetVersion(teamID, appID, fromVersion)
	if errInRetrieveSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return nil, errInRetrieveSnapshot
	}

	// set mainline version
	editVersionAppSnapshot.SetTargetVersion(toVersion)
	editVersionAppSnapshot.SetTriggerMode(snapshotTriggerMode)

	// update old edit version snapshot
	errInUpdateSnapshot := controller.Storage.AppSnapshotStorage.UpdateWholeSnapshot(editVersionAppSnapshot)
	if errInUpdateSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT, "update snapshot failed: "+errInUpdateSnapshot.Error())
		return nil, errInUpdateSnapshot
	}

	// create new edit version snapshot
	newAppSnapShot := model.NewAppSnapshot(teamID, appID, fromVersion, snapshotTriggerMode)
	newAppSnapShot.SetTriggerModeAuto()

	// storage new edit version snapshot
	_, errInCreateSnapshot := controller.Storage.AppSnapshotStorage.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "create snapshot failed: "+errInCreateSnapshot.Error())
		return nil, errInCreateSnapshot
	}

	return newAppSnapShot, nil
}

func (controller *Controller) GetTargetVersionFullApp(c *gin.Context, teamID int, appID int, version int, getPublicApp bool) (*model.FullAppForExport, error) {
	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app full data error: "+errInRetrieveApp.Error())
		return nil, errInRetrieveApp
	}

	// not published to marketplace, check permission
	if getPublicApp && !app.IsPublic() {
		errInGetPublicApp := errors.New("can not access this app")
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app full data error: "+errInGetPublicApp.Error())
		return nil, errInGetPublicApp
	}

	// set for auto-version
	if version == model.APP_AUTO_MAINLINE_VERSION {
		version = app.MainlineVersion
	}
	if version == model.APP_AUTO_RELEASE_VERSION {
		version = app.ReleaseVersion
	}

	// form editor object field appForExport, We need:
	//  -> AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// get all modifier user ids from all apps
	allUserIDs := model.ExtractAllEditorIDFromApps([]*model.App{app})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return nil, errInGetMultiUserInfo
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//  -> Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field actions
	actions, errInRetrieveActions := controller.Storage.ActionStorage.RetrieveActionsByTeamIDAppIDAndVersion(teamID, appID, version)

	// ok, we have no actions for this app
	if errInRetrieveActions != nil {
		actions = []*model.Action{}
	}

	// build actions for expost
	actionsForExport := make([]*model.ActionForExport, 0)
	for _, action := range actions {
		actionForExport := model.NewActionForExport(action)
		// append remote virtual resource
		if actionForExport.Type == resourcelist.TYPE_AI_AGENT {
			api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
			if errInNewAPI != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInNewAPI.Error())
				return nil, errInNewAPI
			}
			aiAgent, _ := api.GetAIAgent(actionForExport.ExportResourceIDInInt())
			actionForExport.AppendVirtualResourceToTemplate(aiAgent)
		}
		actionsForExport = append(actionsForExport, actionForExport)
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//  -> Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field components
	treeStateComponents, errInRetrieveTreeStateComponents := controller.Storage.TreeStateStorage.RetrieveTreeStatesByApp(teamID, appID, model.TREE_STATE_TYPE_COMPONENTS, version)
	fmt.Printf("[DUMP] treeStateComponents: %+v\n", treeStateComponents)
	if errInRetrieveTreeStateComponents != nil {
		fmt.Printf("[DUMP] errInRetrieveTreeStateComponents: %+v\n", errInRetrieveTreeStateComponents)
		treeStateComponents = []*model.TreeState{}
	}
	treeStateLT := model.BuildTreeStateLookupTable(treeStateComponents)
	rootOfTreeState := model.PickUpTreeStatesRootNode(treeStateComponents)
	componentTree, errInBuildComponentsTree := model.BuildComponentTree(rootOfTreeState, treeStateLT, nil, nil)
	if errInBuildComponentsTree != nil {
		fmt.Printf("[DUMP] errInBuildComponentsTree: %+v\n", errInBuildComponentsTree)
	}
	fmt.Printf("[DUMP] componentTree: %+v\n", componentTree)

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//  -> DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field dependenciesState
	dependenciesState := map[string][]string{}
	dependenciesKVStates, errInRetrieveDependenciesKVStates := controller.Storage.KVStateStorage.RetrieveKVStatesByApp(teamID, appID, model.KV_STATE_TYPE_DEPENDENCIES, version)
	if errInRetrieveDependenciesKVStates != nil {
		dependenciesKVStates = []*model.KVState{}
	}
	for _, dependency := range dependenciesKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dependency.Value), &revMsg)
		dependenciesState[dependency.Key] = revMsg // value convert to []string
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//  -> DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field dragShadowState
	dragShadowState := map[string]interface{}{}
	dragShadowKVStates, errInRetrieveDragShadowKVStates := controller.Storage.KVStateStorage.RetrieveKVStatesByApp(teamID, appID, model.KV_STATE_TYPE_DRAG_SHADOW, version)
	if errInRetrieveDragShadowKVStates != nil {
		dragShadowKVStates = []*model.KVState{}
	}
	for _, dragShadow := range dragShadowKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dragShadow.Value), &revMsg)
		dragShadowState[dragShadow.Key] = revMsg // value convert to []string
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//  -> DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field dottedLineSquareState
	dottedLineSquareState := map[string]interface{}{}
	dottedLineSquareKVStates, errInRetrieveDottedLineSquareKVStates := controller.Storage.KVStateStorage.RetrieveKVStatesByApp(teamID, appID, model.KV_STATE_TYPE_DOTTED_LINE_SQUARE, version)
	if errInRetrieveDottedLineSquareKVStates != nil {
		dottedLineSquareKVStates = []*model.KVState{}
	}
	for _, dottedLineSquare := range dottedLineSquareKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dottedLineSquare.Value), &revMsg)
		dottedLineSquareState[dottedLineSquare.Key] = revMsg // value convert to []string
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//  -> DisplayNameState      which is: []string

	// form editor object field displayNameState
	displayNameSetStates, errInRetrieveDisplayNameSetState := controller.Storage.SetStateStorage.RetrieveSetStatesByTeamIDAppIDAndVersion(teamID, appID, model.SET_STATE_TYPE_DISPLAY_NAME, version)
	if errInRetrieveDisplayNameSetState != nil {
		displayNameSetStates = []*model.SetState{}
	}
	displayNameState := make([]string, 0, len(displayNameSetStates))
	for _, displayName := range displayNameSetStates {
		displayNameState = append(displayNameState, displayName.Value)
	}

	// and last one, get app for export with full config info
	appForExport := model.NewAppForExportWithFullConfigInfo(app, usersLT, treeStateComponents, actions)

	// finally, make a brand new editor object
	fullAppForExport := model.NewFullAppForExport(appForExport, actionsForExport, componentTree, dependenciesState, dragShadowState, dottedLineSquareState, displayNameState)

	// feedback
	return fullAppForExport, nil
}

func (controller *Controller) GetTargetVersionFullAppByAppID(c *gin.Context, appID int, version int) (*model.FullAppForExport, error) {
	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveByID(appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app full data error: "+errInRetrieveApp.Error())
		return nil, errInRetrieveApp
	}
	teamID := app.ExportTeamID()

	// set for auto-version
	if version == model.APP_AUTO_MAINLINE_VERSION {
		version = app.MainlineVersion
	}
	if version == model.APP_AUTO_RELEASE_VERSION {
		version = app.ReleaseVersion
	}

	// form editor object field appForExport, We need:
	//  -> AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// get all modifier user ids from all apps
	allUserIDs := model.ExtractAllEditorIDFromApps([]*model.App{app})

	// fet all user id mapped user info, and build user info lookup table
	usersLT, errInGetMultiUserInfo := datacontrol.GetMultiUserInfo(allUserIDs)
	if errInGetMultiUserInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_USER, "get user info failed: "+errInGetMultiUserInfo.Error())
		return nil, errInGetMultiUserInfo
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//  -> Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field actions
	actions, errInRetrieveActions := controller.Storage.ActionStorage.RetrieveActionsByTeamIDAppIDAndVersion(teamID, appID, version)

	// ok, we have no actions for this app
	if errInRetrieveActions != nil {
		actions = []*model.Action{}
	}

	// build actions for expost
	actionsForExport := make([]*model.ActionForExport, 0)
	for _, action := range actions {
		actionForExport := model.NewActionForExport(action)
		// append remote virtual resource
		if actionForExport.Type == resourcelist.TYPE_AI_AGENT {
			api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
			if errInNewAPI != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInNewAPI.Error())
				return nil, errInNewAPI
			}
			aiAgent, errInGetAIAgent := api.GetAIAgent(actionForExport.ExportResourceIDInInt())
			if errInGetAIAgent != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInGetAIAgent.Error())
				return nil, errInGetAIAgent
			}
			actionForExport.AppendVirtualResourceToTemplate(aiAgent)
		}
		actionsForExport = append(actionsForExport, actionForExport)
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//  -> Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field components
	treeStateComponents, errInRetrieveTreeStateComponents := controller.Storage.TreeStateStorage.RetrieveTreeStatesByApp(teamID, appID, model.TREE_STATE_TYPE_COMPONENTS, version)
	if errInRetrieveTreeStateComponents != nil {
		treeStateComponents = []*model.TreeState{}
	}
	treeStateLT := model.BuildTreeStateLookupTable(treeStateComponents)
	rootOfTreeState := model.PickUpTreeStatesRootNode(treeStateComponents)
	componentTree, _ := model.BuildComponentTree(rootOfTreeState, treeStateLT, nil, nil)

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//  -> DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field dependenciesState
	dependenciesState := map[string][]string{}
	dependenciesKVStates, errInRetrieveDependenciesKVStates := controller.Storage.KVStateStorage.RetrieveKVStatesByApp(teamID, appID, model.KV_STATE_TYPE_DEPENDENCIES, version)
	if errInRetrieveDependenciesKVStates != nil {
		dependenciesKVStates = []*model.KVState{}
	}
	for _, dependency := range dependenciesKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dependency.Value), &revMsg)
		dependenciesState[dependency.Key] = revMsg // value convert to []string
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//  -> DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field dragShadowState
	dragShadowState := map[string]interface{}{}
	dragShadowKVStates, errInRetrieveDragShadowKVStates := controller.Storage.KVStateStorage.RetrieveKVStatesByApp(teamID, appID, model.KV_STATE_TYPE_DRAG_SHADOW, version)
	if errInRetrieveDragShadowKVStates != nil {
		dragShadowKVStates = []*model.KVState{}
	}
	for _, dragShadow := range dragShadowKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dragShadow.Value), &revMsg)
		dragShadowState[dragShadow.Key] = revMsg // value convert to []string
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//  -> DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field dottedLineSquareState
	dottedLineSquareState := map[string]interface{}{}
	dottedLineSquareKVStates, errInRetrieveDottedLineSquareKVStates := controller.Storage.KVStateStorage.RetrieveKVStatesByApp(teamID, appID, model.KV_STATE_TYPE_DOTTED_LINE_SQUARE, version)
	if errInRetrieveDottedLineSquareKVStates != nil {
		dottedLineSquareKVStates = []*model.KVState{}
	}
	for _, dottedLineSquare := range dottedLineSquareKVStates {
		var revMsg []string
		json.Unmarshal([]byte(dottedLineSquare.Value), &revMsg)
		dottedLineSquareState[dottedLineSquare.Key] = revMsg // value convert to []string
	}

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//     Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//  -> DisplayNameState      which is: []string

	// form editor object field displayNameState
	displayNameSetStates, errInRetrieveDisplayNameSetState := controller.Storage.SetStateStorage.RetrieveSetStatesByTeamIDAppIDAndVersion(teamID, appID, model.SET_STATE_TYPE_DISPLAY_NAME, version)
	if errInRetrieveDisplayNameSetState != nil {
		displayNameSetStates = []*model.SetState{}
	}
	displayNameState := make([]string, 0, len(displayNameSetStates))
	for _, displayName := range displayNameSetStates {
		displayNameState = append(displayNameState, displayName.Value)
	}

	// and last one, get app for export with full config info
	appForExport := model.NewAppForExportWithFullConfigInfo(app, usersLT, treeStateComponents, actions)

	// finally, make a brand new editor object
	fullAppForExport := model.NewFullAppForExport(appForExport, actionsForExport, componentTree, dependenciesState, dragShadowState, dottedLineSquareState, displayNameState)

	// feedback
	return fullAppForExport, nil
}

func (controller *Controller) BuildComponentTree(app *model.App, parentNodeID int, componentNodeTree *model.ComponentNode) error {
	// convert ComponentNode to TreeState
	currentNode, errInNewCurrentNode := model.NewTreeStateByAppAndComponentState(app, model.TREE_STATE_TYPE_COMPONENTS, componentNodeTree)
	if errInNewCurrentNode != nil {
		return errInNewCurrentNode
	}

	parentTreeState := model.NewTreeState()
	isSummitNode := true

	// get parentNode
	if parentNodeID != model.TREE_STATE_SUMMIT_ID || currentNode.ParentNode == model.TREE_STATE_SUMMIT_NAME { // parentNode is in database
		isSummitNode = false
		var errInRetrieveTreeStateByID error
		parentTreeState, errInRetrieveTreeStateByID = controller.Storage.TreeStateStorage.RetrieveByID(app.ExportTeamID(), parentNodeID)
		if errInRetrieveTreeStateByID != nil {
			return errInRetrieveTreeStateByID
		}
	} else if componentNodeTree.ParentNode != "" && componentNodeTree.ParentNode != model.TREE_STATE_SUMMIT_NAME { // or parentNode is exist
		isSummitNode = false
		var errInRetrieveTreeStateByApp error
		parentTreeState, errInRetrieveTreeStateByApp = controller.Storage.TreeStateStorage.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, componentNodeTree.ParentNode)
		if errInRetrieveTreeStateByApp != nil {
			return errInRetrieveTreeStateByApp
		}
	}

	// no parentNode, currentNode is tree summit
	if isSummitNode && currentNode.Name != model.TREE_STATE_SUMMIT_NAME {
		// get root node
		var errInRetrieveTreeStateByApp error
		parentTreeState, errInRetrieveTreeStateByApp = controller.Storage.TreeStateStorage.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, model.TREE_STATE_SUMMIT_NAME)
		if errInRetrieveTreeStateByApp != nil {
			return errInRetrieveTreeStateByApp
		}
	}
	currentNode.ParentNodeRefID = parentTreeState.ID

	// storage currentNode to database and get id
	_, errInCreateTreeState := controller.Storage.TreeStateStorage.Create(currentNode)
	if errInCreateTreeState != nil {
		return errInCreateTreeState
	}

	// fill parentNode.ChildrenNodeRefIDs with currentNode.ID (when current node is not root)
	if currentNode.Name != model.TREE_STATE_SUMMIT_NAME {
		parentTreeState.AppendChildrenNodeRefIDs([]int{currentNode.ID})
		// update parentNode
		errInUpdateParentNode := controller.Storage.TreeStateStorage.Update(parentTreeState)
		if errInUpdateParentNode != nil {
			return errInUpdateParentNode
		}
	}

	// ok, continue to process currentNode.ChildrenNode
	for _, childrenComponentNode := range componentNodeTree.ChildrenNode {
		errInBuildComponentTree := controller.BuildComponentTree(app, currentNode.ID, childrenComponentNode)
		if errInBuildComponentTree != nil {
			return errInBuildComponentTree
		}
	}
	return nil
}
