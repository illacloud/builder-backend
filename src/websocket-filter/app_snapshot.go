package filter

import (
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/websocket"
)

func TakeSnapshot(hub *websocket.Hub, message *websocket.Message) error {
	currentClient, _ := hub.Clients[message.ClientID]
	teamID := currentClient.TeamID
	appID := currentClient.APPID
	userID := currentClient.MappedUserID

	// fetch app
	app, errInRetrieveApp := hub.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		return errInRetrieveApp
	}

	// config app
	treeStateLatestVersion, _ := hub.Storage.TreeStateStorage.RetrieveTreeStatesLatestVersion(teamID, appID)
	app.SyncMainlineVersionWithTreeStateLatestVersion(treeStateLatestVersion)
	app.BumpMainlineVersion()

	// do snapshot for app following components and actions
	errInSnapshotTreeState := SnapshotTreeState(hub, teamID, appID, app.ExportMainlineVersion())
	if errInSnapshotTreeState != nil {
		return errInSnapshotTreeState
	}
	errInSnapshotKVState := SnapshotKVState(hub, teamID, appID, app.ExportMainlineVersion())
	if errInSnapshotKVState != nil {
		return errInSnapshotKVState
	}
	errInSnapshotSetState := SnapshotSetState(hub, teamID, appID, app.ExportMainlineVersion())
	if errInSnapshotSetState != nil {
		return errInSnapshotSetState
	}
	errInSnapshotAction := SnapshotAction(hub, teamID, appID, app.ExportMainlineVersion())
	if errInSnapshotAction != nil {
		return errInSnapshotAction
	}

	// save snapshot
	_, errInTakeSnapshot := SaveAppSnapshot(hub, teamID, appID, userID, app.ExportMainlineVersion(), model.SNAPSHOT_TRIGGER_MODE_AUTO)
	if errInTakeSnapshot != nil {
		return errInTakeSnapshot
	}

	// update app for version bump
	errInUpdateApp := hub.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		return errInUpdateApp
	}

	// ok
	return nil
}

func SnapshotTreeState(hub *websocket.Hub, teamID int, appID int, appMainLineVersion int) error {
	return DuplicateTreeStateByVersion(hub, teamID, appID, model.APP_EDIT_VERSION, appMainLineVersion)
}

func SnapshotKVState(hub *websocket.Hub, teamID int, appID int, appMainLineVersion int) error {
	return DuplicateKVStateByVersion(hub, teamID, appID, model.APP_EDIT_VERSION, appMainLineVersion)
}

func SnapshotSetState(hub *websocket.Hub, teamID int, appID int, appMainLineVersion int) error {
	return DuplicateSetStateByVersion(hub, teamID, appID, model.APP_EDIT_VERSION, appMainLineVersion)
}

func SnapshotAction(hub *websocket.Hub, teamID int, appID int, mainlineVersion int) error {
	return DuplicateActionByVersion(hub, teamID, appID, model.APP_EDIT_VERSION, mainlineVersion)
}

// recover edit version treestate to target version (coby target version data to edit version)
func DuplicateTreeStateByVersion(hub *websocket.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	// get from version tree state from database
	treestates, errInRetrieveTreeState := hub.Storage.TreeStateStorage.RetrieveTreeStatesByTeamIDAppIDAndVersion(teamID, appID, fromVersion)
	if errInRetrieveTreeState != nil {
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
		newID, errInCreateApp := hub.Storage.TreeStateStorage.Create(treestate)
		if errInCreateApp != nil {
			return errInCreateApp
		}
		oldID := oldIDMap[i]
		releaseIDMap[oldID] = newID
	}

	// update children node ids
	for _, treestate := range treestates {
		treestate.RemapChildrenNodeRefIDs(releaseIDMap)
		treestate.SetParentNodeRefID(releaseIDMap[treestate.ParentNodeRefID])
		errInUpdateTreeState := hub.Storage.TreeStateStorage.Update(treestate)
		if errInUpdateTreeState != nil {
			return errInUpdateTreeState
		}
	}

	return nil
}

func DuplicateKVStateByVersion(hub *websocket.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	kvstates, errInRetrieveKVState := hub.Storage.KVStateStorage.RetrieveKVStatesByTeamIDAppIDAndVersion(teamID, appID, fromVersion)
	if errInRetrieveKVState != nil {
		return errInRetrieveKVState
	}

	// set version as mainline version
	for serial, _ := range kvstates {
		kvstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		errInCreateKVState := hub.Storage.KVStateStorage.Create(kvstate)
		if errInCreateKVState != nil {
			return errInCreateKVState
		}
	}
	return nil
}

func DuplicateSetStateByVersion(hub *websocket.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	setstates, errInRetrieveSetState := hub.Storage.SetStateStorage.RetrieveSetStatesByTeamIDAppIDAndVersion(teamID, appID, model.SET_STATE_TYPE_DISPLAY_NAME, fromVersion)
	if errInRetrieveSetState != nil {
		return errInRetrieveSetState
	}

	// update some fields
	for serial, _ := range setstates {
		setstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, setstate := range setstates {
		errInCreateSetState := hub.Storage.SetStateStorage.Create(setstate)
		if errInCreateSetState != nil {
			return errInCreateSetState
		}
	}
	return nil
}

func DuplicateActionByVersion(hub *websocket.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	actions, errinRetrieveAction := hub.Storage.ActionStorage.RetrieveActionsByTeamIDAppIDAndVersion(teamID, appID, fromVersion)
	if errinRetrieveAction != nil {
		return errinRetrieveAction
	}

	// set version as mainline version
	for serial, _ := range actions {
		actions[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, action := range actions {
		_, errInCreateAction := hub.Storage.ActionStorage.Create(action)
		if errInCreateAction != nil {
			return errInCreateAction
		}
	}
	return nil
}

func SaveAppSnapshot(hub *websocket.Hub, teamID int, appID int, userID int, mainlineVersion int, snapshotTriggerMode int) (*model.AppSnapshot, error) {
	return SaveAppSnapshotByVersion(hub, teamID, appID, userID, model.APP_EDIT_VERSION, mainlineVersion, snapshotTriggerMode)
}

// SaveAppSnapshotByVersion() method do following process:
// - get current version snapshot
// - set it to target version
// - save it
// - create new empty snapshot for current version
func SaveAppSnapshotByVersion(hub *websocket.Hub, teamID int, appID int, userID int, fromVersion int, toVersion int, snapshotTriggerMode int) (*model.AppSnapshot, error) {
	// retrieve app mainline version snapshot
	editVersionAppSnapshot, errInRetrieveSnapshot := hub.Storage.AppSnapshotStorage.RetrieveByTeamIDAppIDAndTargetVersion(teamID, appID, fromVersion)
	if errInRetrieveSnapshot != nil {
		return nil, errInRetrieveSnapshot
	}

	// set mainline version
	editVersionAppSnapshot.SetTargetVersion(toVersion)

	// update old edit version snapshot
	errInUpdateSnapshot := hub.Storage.AppSnapshotStorage.UpdateWholeSnapshot(editVersionAppSnapshot)
	if errInUpdateSnapshot != nil {
		return nil, errInUpdateSnapshot
	}

	// create new edit version snapshot
	newAppSnapShot := model.NewAppSnapshot(teamID, appID, fromVersion, snapshotTriggerMode)

	// storage new edit version snapshot
	_, errInCreateSnapshot := hub.Storage.AppSnapshotStorage.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		return nil, errInCreateSnapshot
	}

	return newAppSnapShot, nil
}
