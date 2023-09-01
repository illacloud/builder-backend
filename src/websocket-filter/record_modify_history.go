package filter

import (
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/websocket"
)

func RecordModifyHistory(hub *websocket.Hub, message *websocket.Message, displayNames []string) error {
	// check deploy env
	conf := config.GetInstance()
	if !conf.IsCloudMode() {
		return nil
	}
	// go
	currentClient, _ := hub.Clients[message.ClientID]
	teamID := currentClient.TeamID
	appID := currentClient.APPID
	userID := currentClient.MappedUserID

	// get current edit version app snapshot
	appSnapshot, errInGetSnapshot := hub.Storage.AppSnapshotStorage.RetrieveEditVersion(teamID, appID)
	if errInGetSnapshot != nil {
		currentClient.Feedback(message, websocket.ERROR_CREATE_SNAPSHOT_MIDIFY_HISTORY_FAILED, errInGetSnapshot)
		return errInGetSnapshot
	}

	// new modify history
	for _, displayName := range displayNames {
		broadcastType := ""
		var broadcastPayload interface{}
		if message.Broadcast != nil {
			broadcastType = message.Broadcast.Type
			broadcastPayload = message.Broadcast.Payload
		}
		modifyHistoryRecord := model.NewAppModifyHistory(message.Signal, message.Target, displayName, broadcastType, broadcastPayload, userID)
		appSnapshot.PushModifyHistory(modifyHistoryRecord)
	}

	// update app snapshot
	errInUpdateSnapshot := hub.Storage.AppSnapshotStorage.UpdateWholeSnapshot(appSnapshot)
	if errInUpdateSnapshot != nil {
		currentClient.Feedback(message, websocket.ERROR_UPDATE_SNAPSHOT_MIDIFY_HISTORY_FAILED, errInUpdateSnapshot)
		return errInUpdateSnapshot
	}

	// check if app snapshot need archive
	if !appSnapshot.DoesActiveSnapshotNeedArchive() {
		return nil
	}

	// ok, archive app snapshot
	TakeSnapshot(hub, message)

	return nil
}
