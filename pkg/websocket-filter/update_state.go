package filter

import (
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalUpdateState(hub *ws.Hub, message *ws.Message) error {

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.APPID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case ws.TARGET_NOTNING:
		return nil
	case ws.TARGET_COMPONENTS:
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
				currentClient.Feedback(message, ERROR_UPDATE_STATE_FAILED, err)
				return err
			}
		}

		// feedback currentClient
		currentClient.Feedback(message, ERROR_UPDATE_STATE_OK)

		// feedback otherClient
		hub.BroadcastToOtherClients(message, currentClient)
	case ws.TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case ws.TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case ws.TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// update K-V State
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType

			// update
			if err := hub.KVStateServiceImpl.UpdateKVStateByKey(apprefid, &kvstatedto); err != nil {
				currentClient.Feedback(message, ERROR_UPDATE_STATE_FAILED, err)
				return err
			}
		}
	case ws.TARGET_DISPLAY_NAME:
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
				currentClient.Feedback(message, ERROR_CREATE_STATE_FAILED, err)
				return err
			}
		}
	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	currentClient.Feedback(message, ERROR_UPDATE_STATE_OK, nil)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)

	return nil
}
