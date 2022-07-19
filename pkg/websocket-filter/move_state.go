package filter

import (
	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalMoveState(hub *websocket.Hub, message *websocket.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		apprefid := currentClient.RoomID
		for _, v := range message.Payload {
			var nowNode state.TreeStateDto
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS

			if err := hub.TreeStateServiceImpl.MoveTreeStateNode(apprefid, &nowNode); err != nil {
				FeedbackCurrentClient(message, currentClient, ERROR_MOVE_STATE_FAILED)
				return err
			}
		}

		// feedback currentClient
		FeedbackCurrentClient(message, currentClient, ERROR_MOVE_STATE_OK)

		// feedback otherClient
		BroadcastToOtherClients(hub, message, currentClient)

	case TARGET_DEPENDENCIES:
		fallthrough
	case TARGET_DRAG_SHADOW:
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		FeedbackCurrentClient(message, currentClient, ERROR_CAN_NOT_MOVE_KVSTATE)
		return nil
	case TARGET_DISPLAY_NAME:
		FeedbackCurrentClient(message, currentClient, ERROR_CAN_NOT_MOVE_SETSTATE)
		return nil
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}
	return nil
}
