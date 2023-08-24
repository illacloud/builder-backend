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
	"log"

	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/util/illaresourcemanagerbackendsdk"
	"github.com/illacloud/builder-backend/internal/util/resourcelist"
	"github.com/illacloud/builder-backend/src/model"
	repository "github.com/illacloud/builder-backend/src/model"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) SnapshotTreeState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	return controller.DuplicateTreeStateByVersion(c, teamID, appID, model.APP_EDIT_VERSION, appMainLineVersion)
}

func (controller *Controller) SnapshotKVState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	return controller.DuplicateKVStateByVersion(c, teamID, appID, model.APP_EDIT_VERSION, appMainLineVersion)
}

func (controller *Controller) SnapshotSetState(c *gin.Context, teamID int, appID int, appMainLineVersion int) error {
	return controller.DuplicateSetStateByVersion(c, teamID, appID, model.APP_EDIT_VERSION, appMainLineVersion)
}

func (controller *Controller) SnapshotAction(c *gin.Context, teamID int, appID int, mainlineVersion int) error {
	return controller.DuplicateActionByVersion(c, teamID, appID, model.APP_EDIT_VERSION, mainlineVersion)
}

// recover edit version treestate to target version (coby target version data to edit version)
func (controller *Controller) DuplicateTreeStateByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	// get from version tree state from database
	treestates, errInRetrieveTreeState := controller.Storage.TreeStateStorage.RetrieveAllTypeTreeStatesByApp(teamID, appID, fromVersion)
	if errInRetrieveTreeState != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get tree state failed: "+errInRetrieveTreeState.Error())
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
		log.Printf("[DUMP] DuplicateTreeStateByVersion: treestate.Name: %s, treestate.TeamID: %d, treestate.AppRefID: %d, , treestate.Version: %d\n", treestate.Name, treestate.TeamID, treestate.AppRefID, treestate.Version)
		newID, errInCreateApp := controller.Storage.TreeStateStorage.Create(treestate)
		if errInCreateApp != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_APP, "create app failed: "+errInCreateApp.Error())
			return errInCreateApp
		}
		oldID := oldIDMap[i]
		releaseIDMap[oldID] = newID
	}

	// update children node ids
	for _, treestate := range treestates {
		treestate.RemapChildrenNodeRefIDs(releaseIDMap)
		treestate.SetParentNodeRefID(releaseIDMap[treestate.ParentNodeRefID])
		errInUpdateTreeState := controller.Storage.TreeStateStorage.Update(treestate)
		if errInUpdateTreeState != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_TREE_STATE, "update tree state failed: "+errInUpdateTreeState.Error())
			return errInUpdateTreeState
		}
	}

	return nil
}

func (controller *Controller) DuplicateKVStateByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	kvstates, errInRetrieveKVState := controller.Storage.KVStateStorage.RetrieveAllTypeKVStatesByApp(teamID, appID, fromVersion)
	if errInRetrieveKVState != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get kv state failed: "+errInRetrieveKVState.Error())
		return errInRetrieveKVState
	}

	// set version as mainline version
	for serial, _ := range kvstates {
		kvstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		log.Printf("[DUMP] DuplicateKVStateByVersion: kvstate.StateType: %d, kvstate.TeamID: %d, kvstate.AppRefID: %d, , kvstate.Version: %d\n", kvstate.StateType, kvstate.TeamID, kvstate.AppRefID, kvstate.Version)
		errInCreateKVState := controller.Storage.KVStateStorage.Create(kvstate)
		if errInCreateKVState != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_STATE, "create kv state failed: "+errInCreateKVState.Error())
			return errInCreateKVState
		}
	}
	return nil
}

func (controller *Controller) DuplicateSetStateByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	setstates, errInRetrieveSetState := controller.Storage.SetStateStorage.RetrieveSetStatesByApp(teamID, appID, model.SET_STATE_TYPE_DISPLAY_NAME, fromVersion)
	if errInRetrieveSetState != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_STATE, "get set state failed: "+errInRetrieveSetState.Error())
		return errInRetrieveSetState
	}

	// update some fields
	for serial, _ := range setstates {
		setstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, setstate := range setstates {
		log.Printf("[DUMP] DuplicateSetStateByVersion: setstate.StateType: %d, setstate.TeamID: %d, setstate.AppRefID: %d, , setstate.Version: %d\n", setstate.StateType, setstate.TeamID, setstate.AppRefID, setstate.Version)
		errInCreateSetState := controller.Storage.SetStateStorage.Create(setstate)
		if errInCreateSetState != nil {
			controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_STATE, "create set state failed: "+errInCreateSetState.Error())
			return errInCreateSetState
		}
	}
	return nil
}

func (controller *Controller) DuplicateActionByVersion(c *gin.Context, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	actions, errinRetrieveAction := controller.Storage.ActionStorage.RetrieveActionsByAppVersion(teamID, appID, fromVersion)
	if errinRetrieveAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action failed: "+errinRetrieveAction.Error())
		return errinRetrieveAction
	}

	// set version as mainline version
	for serial, _ := range actions {
		actions[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, action := range actions {
		log.Printf("[DUMP] DuplicateActionByVersion: action.Name: %s, action.TeamID: %d, action.AppRefID: %d, , action.Version: %d\n", action.Name, action.TeamID, action.App, action.Version)
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
	_, errInCreateSnapshot := controller.AppSnapshotmodel.Create(newAppSnapShot)
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
	editVersionAppSnapshot, errInRetrieveSnapshot := controller.AppSnapshotmodel.RetrieveByTeamIDAppIDAndTargetVersion(teamID, appID, fromVersion)
	if errInRetrieveSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_SNAPSHOT, "get snapshot failed: "+errInRetrieveSnapshot.Error())
		return nil, errInRetrieveSnapshot
	}

	// set mainline version
	editVersionAppSnapshot.SetTargetVersion(toVersion)
	editVersionAppSnapshot.SetTriggerMode(snapshotTriggerMode)

	// update old edit version snapshot
	errInUpdateSnapshot := controller.AppSnapshotmodel.UpdateWholeSnapshot(editVersionAppSnapshot)
	if errInUpdateSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_SNAPSHOT, "update snapshot failed: "+errInUpdateSnapshot.Error())
		return nil, errInUpdateSnapshot
	}

	// create new edit version snapshot
	newAppSnapShot := model.NewAppSnapshot(teamID, appID, fromVersion, snapshotTriggerMode)
	newAppSnapShot.SetTriggerModeAuto()

	// storage new edit version snapshot
	_, errInCreateSnapshot := controller.AppSnapshotmodel.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_SNAPSHOT, "create snapshot failed: "+errInCreateSnapshot.Error())
		return nil, errInCreateSnapshot
	}

	return newAppSnapShot, nil
}

func (controller *Controller) GetTargetVersionFullApp(c *gin.Context, teamID int, appID int, version int) (*model.NewFullAppForExport, error) {
	// fetch app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app full data error: "+errInRetrieveApp.Error())
		return nil, errInRetrieveApp
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

	appForExport := model.NewAppForExport(app, usersLT)

	// form editor object field appForExport, We need:
	//     AppInfo               which is: *AppForExport
	//  -> Actions               which is: []*ActionForExport
	//     Components            which is: *ComponentNode
	//     DependenciesState     which is: map[string][]string
	//     DragShadowState       which is: map[string]interface{}
	//     DottedLineSquareState which is: map[string]interface{}
	//     DisplayNameState      which is: []string

	// form editor object field actions
	actions, errInRetrieveActions := controller.Storage.ActionStorage.RetrieveActionsByAppVersion(teamID, appID, version)

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
			api, errInNewAPI := illaresourcemanagerbackendsdk.NewIllaResourceManagerRestAPI()
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
	componentTree, _ := model.BuildComponentTree(rootOfTreeState, treeStateLT, nil)

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
	displayNameSetStates, errInRetrieveDisplayNameSetState := controller.Storage.SetStateStorage.RetrieveSetStatesByApp(teamID, appID, model.SET_STATE_TYPE_DISPLAY_NAME, version)
	if errInRetrieveDisplayNameSetState != nil {
		displayNameSetStates = []*model.SetState{}
	}
	displayNameState := make([]string, 0, len(displayNameSetStates))
	for _, displayName := range displayNameSetStates {
		displayNameState = append(displayNameState, displayName.Value)
	}

	// finally, make a brand new editor object
	fullAppForExport := model.NewFullAppForExport(appForExport, actionsForExport, componentTree, dependenciesState, dragShadowState, dottedLineSquareState, displayNameState)

	// feedback
	return fullAppForExport, nil
}

func (controller *Controller) CreateComponentTree(app *repository.App, parentNodeID int, componentNodeTree *repository.ComponentNode) error {
	// summit node
	if parentNodeID == 0 {
		parentNodeID = repository.TREE_STATE_SUMMIT_ID
	}

	// convert ComponentNode to TreeState
	currentNode := NewTreeStateDto()
	currentNode.SetTeamID(app.ExportTeamID())
	currentNode.InitUID()
	currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS)
	var err error
	if currentNode, err = impl.NewTreeStateByComponentState(app, componentNodeTree); err != nil {
		return err
	}

	// get parentNode
	parentTreeState := repository.NewTreeState()
	isSummitNode := true
	if parentNodeID != 0 || currentNode.ParentNode == repository.TREE_STATE_SUMMIT_NAME { // parentNode is in database
		isSummitNode = false
		if parentTreeState, err = impl.treestateRepository.RetrieveByID(app.ExportTeamID(), parentNodeID); err != nil {
			return err
		}
	} else if componentNodeTree.ParentNode != "" && componentNodeTree.ParentNode != repository.TREE_STATE_SUMMIT_NAME { // or parentNode is exist
		isSummitNode = false
		if parentTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, componentNodeTree.ParentNode); err != nil {
			return err
		}
	}

	// no parentNode, currentNode is tree summit
	if isSummitNode && currentNode.Name != repository.TREE_STATE_SUMMIT_NAME {

		// get root node
		if parentTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, repository.TREE_STATE_SUMMIT_NAME); err != nil {
			return err
		}
	}
	currentNode.ParentNodeRefID = parentTreeState.ID

	// insert currentNode and get id
	treeStateDtoInDB := &TreeStateDto{}
	if treeStateDtoInDB, err = impl.CreateTreeState(currentNode); err != nil {
		return err
	}
	currentNode.ID = treeStateDtoInDB.ID

	// fill currentNode id into parentNode.ChildrenNodeRefIDs
	if currentNode.Name != repository.TREE_STATE_SUMMIT_NAME {

		parentTreeState.AppendChildrenNodeRefIDs(currentNode.ID)

		// save parentNode
		if err = impl.treestateRepository.Update(parentTreeState); err != nil {
			return err
		}
	}

	// create currentNode.ChildrenNode
	for _, childrenComponentNode := range componentNodeTree.ChildrenNode {
		if err := impl.CreateComponentTree(app, currentNode.ID, childrenComponentNode); err != nil {
			return err
		}
	}
	return nil
}
