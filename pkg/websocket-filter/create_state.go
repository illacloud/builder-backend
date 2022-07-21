package filter

import (
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalCreateState(hub *ws.Hub, message *ws.Message) error {
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
		// build component tree from json

		summitnodeid := repository.TREE_STATE_SUMMIT_ID

		for _, v := range message.Payload {
			var componenttree *repository.ComponentNode
			componenttree = repository.ConstructComponentNodeByMap(v)

			if err := hub.TreeStateServiceImpl.CreateComponentTree(apprefid, summitnodeid, componenttree); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
		}
	case ws.TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case ws.TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case ws.TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		// create k-v state

		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType

			if _, err := hub.KVStateServiceImpl.CreateKVState(kvstatedto); err != nil {
				currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
				return err
			}
		}
	case ws.TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		// create set state

		for _, v := range message.Payload {
			// resolve payload
			dns, err := repository.ConstructDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// save state
			for _, displayName := range dns {
				var setStateDto state.SetStateDto
				setStateDto.ConstructByValue(displayName)
				setStateDto.StateType = stateType
				// create state
				if _, err := hub.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
					currentClient.Feedback(message, ws.ERROR_CREATE_STATE_FAILED, err)
					return err
				}
			}
		}
	case ws.TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case ws.TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}
	// feedback currentClient
	currentClient.Feedback(message, ws.ERROR_CREATE_STATE_OK, nil)

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)
	return nil
}
