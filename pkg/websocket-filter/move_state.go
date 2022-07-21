package filter

import (
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalMoveState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.APPID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case ws.TARGET_NOTNING:
		return nil
	case ws.TARGET_COMPONENTS:
		apprefid := currentClient.APPID
		for _, v := range message.Payload {
			var nowNode state.TreeStateDto
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS

			if err := hub.TreeStateServiceImpl.MoveTreeStateNode(apprefid, &nowNode); err != nil {
				ws.FeedbackCurrentClient(message, currentClient, ERROR_MOVE_STATE_FAILED)
				return err
			}
		}

		// feedback currentClient
		ws.FeedbackCurrentClient(message, currentClient, ERROR_MOVE_STATE_OK)

		// feedback otherClient
		ws.BroadcastToOtherClients(hub, message, currentClient)

	case ws.TARGET_DEPENDENCIES:
		fallthrough
	case ws.TARGET_DRAG_SHADOW:
		fallthrough
	case ws.TARGET_DOTTED_LINE_SQUARE:
		ws.FeedbackCurrentClient(message, currentClient, ERROR_CAN_NOT_MOVE_KVSTATE)
		return nil
	case ws.TARGET_DISPLAY_NAME:
		ws.FeedbackCurrentClient(message, currentClient, ERROR_CAN_NOT_MOVE_SETSTATE)
		return nil
	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}
	return nil
}
