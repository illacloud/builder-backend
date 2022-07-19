package filter

import (
	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalUpdateState(hub *websocket.Hub, message *websocket.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		for _, v := range message.Payload {
			// update
			// construct update data
			var nowNode state.TreeStateDto
			componentNode := repository.ConstructComponentNodeByMap(v)

			serializedComponent, err := componentNode.SerializationForDatabase()
			if err != nil {
				return err
			}

			nowNode.Content = string(serializedComponent)
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS

			// update
			if err := hub.TreeStateServiceImpl.UpdateTreeStateNode(apprefid, &nowNode); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
				return err
			}
		}

		// feedback currentClient
		FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_OK)

		// feedback otherClient
		BroadcastToOtherClients(hub, message, currentClient)
	case TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// update K-V State
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType

			// update
			if err := hub.KVStateServiceImpl.UpdateKVStateByKey(apprefid, &kvstatedto); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
				return err
			}
		}
	case TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			// resolve payload
			dnsfu, err := repository.ConstructDisplayNameStateForUpdateByPayload(v)
			if err != nil {
				return err
			}
			// init state dto
			var beforeSetStateDto state.SetStateDto
			var afterSetStateDto state.SetStateDto
			beforeSetStateDto.ConstructByDisplayNameForUpdate(dnsfu)
			beforeSetStateDto.StateType = stateType
			beforeSetStateDto.AppRefID = apprefid
			beforeSetStateDto.Version = repository.APP_EDIT_VERSION
			afterSetStateDto.ConstructByDisplayNameForUpdate(dnsfu)
			// update state
			if err := hub.SetStateServiceImpl.UpdateSetStateByValue(beforeSetStateDto, afterSetStateDto); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
				return err
			}
		}
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_OK)

	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)

	return nil
}
