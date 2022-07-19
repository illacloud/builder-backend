package filter

import (
	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalDeleteState(hub *websocket.Hub, message *websocket.Message) error {

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
			var nowNode state.TreeStateDto
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS

			if err := hub.TreeStateServiceImpl.DeleteTreeStateNodeRecursive(apprefid, &nowNode); err != nil {
				websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_DELETE_STATE_FAILED)
				return err
			}
		}

	case TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// delete k-v state
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType

			if err := hub.KVStateServiceImpl.DeleteKVStateByKey(apprefid, &kvstatedto); err != nil {
				websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_DELETE_STATE_FAILED)
				return err
			}
		}

	case TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		// create dnsplayName state

		for _, v := range message.Payload {
			// resolve payload
			dns, err := repository.ConstructDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// save state
			for _, displayName := range dns {
				// init
				var setStateDto state.SetStateDto
				setStateDto.ConstructByValue(displayName)
				setStateDto.StateType = stateType
				setStateDto.AppRefID = apprefid
				setStateDto.Version = repository.APP_EDIT_VERSION
				// delete state
				if err := hub.SetStateServiceImpl.DeleteSetStateByValue(&setStateDto); err != nil {
					websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_CREATE_STATE_FAILED)
					return err
				}
			}
		}
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_DELETE_STATE_OK)
	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)
	return nil
}
