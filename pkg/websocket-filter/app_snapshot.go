package filter

import (
	"github.com/illacloud/builder-backend/internal/repository"
	ws "github.com/illacloud/builder-backend/internal/websocket"
)

func TakeSnapshot(hub *ws.Hub, message *ws.Message) error {
	currentClient, _ := hub.Clients[message.ClientID]
	teamID := currentClient.TeamID
	appID := currentClient.APPID
	userID := currentClient.MappedUserID

	// fetch app
	app, errInRetrieveApp := hub.AppRepositoryImpl.RetrieveAppByIDAndTeamID(appID, teamID)
	if errInRetrieveApp != nil {
		return errInRetrieveApp
	}

	// config app
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
	_, errInTakeSnapshot := SaveAppSnapshot(hub, teamID, appID, userID, app.ExportMainlineVersion(), repository.SNAPSHOT_TRIGGER_MODE_MANUAL)
	if errInTakeSnapshot != nil {
		return errInTakeSnapshot
	}

	// update app for version bump
	errInUpdateApp := hub.AppRepositoryImpl.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		return errInUpdateApp
	}

	// ok
	return nil
}

func SnapshotTreeState(hub *ws.Hub, teamID int, appID int, appMainLineVersion int) error {
	return DuplicateTreeStateByVersion(hub, teamID, appID, repository.APP_EDIT_VERSION, appMainLineVersion)
}

func SnapshotKVState(hub *ws.Hub, teamID int, appID int, appMainLineVersion int) error {
	return DuplicateKVStateByVersion(hub, teamID, appID, repository.APP_EDIT_VERSION, appMainLineVersion)
}

func SnapshotSetState(hub *ws.Hub, teamID int, appID int, appMainLineVersion int) error {
	return DuplicateSetStateByVersion(hub, teamID, appID, repository.APP_EDIT_VERSION, appMainLineVersion)
}

func SnapshotAction(hub *ws.Hub, teamID int, appID int, mainlineVersion int) error {
	return DuplicateActionByVersion(hub, teamID, appID, repository.APP_EDIT_VERSION, mainlineVersion)
}

// recover edit version treestate to target version (coby target version data to edit version)
func DuplicateTreeStateByVersion(hub *ws.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	// get from version tree state from database
	treestates, errInRetrieveTreeState := hub.TreeStateRepositoryImpl.RetrieveAllTypeTreeStatesByApp(teamID, appID, fromVersion)
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
		newID, errInCreateApp := hub.TreeStateRepositoryImpl.Create(treestate)
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
		errInUpdateTreeState := hub.TreeStateRepositoryImpl.Update(treestate)
		if errInUpdateTreeState != nil {
			return errInUpdateTreeState
		}
	}

	return nil
}

func DuplicateKVStateByVersion(hub *ws.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	kvstates, errInRetrieveKVState := hub.KVStateRepositoryImpl.RetrieveAllTypeKVStatesByApp(teamID, appID, fromVersion)
	if errInRetrieveKVState != nil {
		return errInRetrieveKVState
	}

	// set version as mainline version
	for serial, _ := range kvstates {
		kvstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		errInCreateKVState := hub.KVStateRepositoryImpl.Create(kvstate)
		if errInCreateKVState != nil {
			return errInCreateKVState
		}
	}
	return nil
}

func DuplicateSetStateByVersion(hub *ws.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	setstates, errInRetrieveSetState := hub.SetStateRepositoryImpl.RetrieveSetStatesByApp(teamID, appID, repository.SET_STATE_TYPE_DISPLAY_NAME, fromVersion)
	if errInRetrieveSetState != nil {
		return errInRetrieveSetState
	}

	// update some fields
	for serial, _ := range setstates {
		setstates[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, setstate := range setstates {
		errInCreateSetState := hub.SetStateRepositoryImpl.Create(setstate)
		if errInCreateSetState != nil {
			return errInCreateSetState
		}
	}
	return nil
}

func DuplicateActionByVersion(hub *ws.Hub, teamID int, appID int, fromVersion int, toVersion int) error {
	// get edit version K-V state from database
	actions, errinRetrieveAction := hub.ActionRepositoryImpl.RetrieveActionsByAppVersion(teamID, appID, fromVersion)
	if errinRetrieveAction != nil {
		return errinRetrieveAction
	}

	// set version as mainline version
	for serial, _ := range actions {
		actions[serial].AppendNewVersion(toVersion)
	}

	// and put them to the database as duplicate
	for _, action := range actions {
		_, errInCreateAction := hub.ActionRepositoryImpl.Create(action)
		if errInCreateAction != nil {
			return errInCreateAction
		}
	}
	return nil
}

func SaveAppSnapshot(hub *ws.Hub, teamID int, appID int, userID int, mainlineVersion int, snapshotTriggerMode int) (*repository.AppSnapshot, error) {
	return SaveAppSnapshotByVersion(hub, teamID, appID, userID, repository.APP_EDIT_VERSION, mainlineVersion, snapshotTriggerMode)
}

// SaveAppSnapshotByVersion() method do following process:
// - get current version snapshot
// - set it to target version
// - save it
// - create new empty snapshot for current version
func SaveAppSnapshotByVersion(hub *ws.Hub, teamID int, appID int, userID int, fromVersion int, toVersion int, snapshotTriggerMode int) (*repository.AppSnapshot, error) {
	// retrieve app mainline version snapshot
	editVersionAppSnapshot, errInRetrieveSnapshot := hub.AppSnapshotRepositoryImpl.RetrieveByTeamIDAppIDAndTargetVersion(teamID, appID, fromVersion)
	if errInRetrieveSnapshot != nil {
		return nil, errInRetrieveSnapshot
	}

	// set mainline version
	editVersionAppSnapshot.SetTargetVersion(toVersion)

	// add modify history
	modifyHistoryLog := repository.NewTakeAppSnapshotModifyHistory(userID)
	editVersionAppSnapshot.PushModifyHistory(modifyHistoryLog)

	// update old edit version snapshot
	errInUpdateSnapshot := hub.AppSnapshotRepositoryImpl.UpdateWholeSnapshot(editVersionAppSnapshot)
	if errInUpdateSnapshot != nil {
		return nil, errInUpdateSnapshot
	}

	// create new edit version snapshot
	newAppSnapShot := repository.NewAppSnapshot(teamID, appID, fromVersion, snapshotTriggerMode)

	// storage new edit version snapshot
	_, errInCreateSnapshot := hub.AppSnapshotRepositoryImpl.Create(newAppSnapShot)
	if errInCreateSnapshot != nil {
		return nil, errInCreateSnapshot
	}

	return newAppSnapShot, nil
}
