package filter

import (
	"errors"

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
				currentClient.Feedback(message, ERROR_MOVE_STATE_FAILED, err)
				return err
			}
		}

		// feedback currentClient
		currentClient.Feedback(message, ERROR_MOVE_STATE_OK, nil)

		// feedback otherClient
		hub.BroadcastToOtherClients(message, currentClient)

	case ws.TARGET_DEPENDENCIES:
		fallthrough
	case ws.TARGET_DRAG_SHADOW:
		fallthrough
	case ws.TARGET_DOTTED_LINE_SQUARE:
		err := errors.New("K-V State do not suppory move method.")
		currentClient.Feedback(message, ERROR_CAN_NOT_MOVE_KVSTATE, err)
		return nil
	case ws.TARGET_DISPLAY_NAME:
		err := errors.New("Set State do not suppory move method.")
		currentClient.Feedback(message, ERROR_CAN_NOT_MOVE_SETSTATE, err)
		return nil
	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}
	return nil
}
