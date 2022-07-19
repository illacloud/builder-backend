

package filter

import (
	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalCreateOrUpdateState(hub *websocket.Hub, message *websocket.Message) error {
	

	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	stateType := repository.STATE_TYPE_INVALIED
	apprefid := 
	var appDto app.AppDto
	appDto.ConstructByID(currentClient.RoomID)
	message.RewriteBroadcast()
	// target switch
	switch message.Target {
	case TARGET_NOTNING:
		return nil
	case TARGET_COMPONENTS:
		for _, v := range message.Payload {
			// check if state already in database
			var nowNode state.TreeStateDto
			nowNode.ConstructByMap(v) // set Name
			nowNode.StateType = repository.TREE_STATE_TYPE_COMPONENTS
			isStateExists := hub.TreeStateServiceImpl.IsTreeStateNodeExists(apprefid, &nowNode)
			if !isStateExists {
				// create
				summitnodeid := repository.TREE_STATE_SUMMIT_ID
				var componenttree *repository.ComponentNode
				componenttree = repository.ConstructComponentNodeByMap(v)
				
				if err := hub.TreeStateServiceImpl.CreateComponentTree(apprefid, summitnodeid, componenttree); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
			} else {
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

		}
	case TARGET_DEPENDENCIES:
		stateType = repository.KV_STATE_TYPE_DEPENDENCIES
		fallthrough
	case TARGET_DRAG_SHADOW:
		stateType = repository.KV_STATE_TYPE_DRAG_SHADOW
		fallthrough
	case TARGET_DOTTED_LINE_SQUARE:
		stateType = repository.KV_STATE_TYPE_DOTTED_LINE_SQUARE
		for _, v := range message.Payload {
			// fill KVStateDto
			var kvstatedto state.KVStateDto
			kvstatedto.ConstructByMap(v)
			kvstatedto.StateType = stateType
			

			isStateExists := hub.KVStateServiceImpl.IsKVStateNodeExists(apprefid, &kvstatedto)
			if !isStateExists {
				if _, err := hub.KVStateServiceImpl.CreateKVState(kvstatedto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
			} else {
				// update
				if err := hub.KVStateServiceImpl.UpdateKVStateByKey(apprefid, &kvstatedto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
					return err
				}
			}

		}
	case TARGET_DISPLAY_NAME:
		stateType = repository.KV_STATE_TYPE_DISPLAY_NAME
		for _, v := range message.Payload {
			// resolve payload
			dns, err := repository.ConstructDisplayNameStateByPayload(v)
			if err != nil {
				return err
			}
			// create or update state
			for _, displayName := range dns {
				// checkout
				var setStateDto state.SetStateDto
				var setStateDtoInDB *state.SetStateDto
				var err error
				setStateDto.ConstructByValue(displayName)
				setStateDto.ConstructByType(stateType)
				setStateDto.ConstructByApp(appDto)
				setStateDto.ConstructWithEditVersion()
				// lookup state
				if setStateDtoInDB, err = hub.SetStateServiceImpl.GetByValue(setStateDto); err != nil {
					FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
					return err
				}
				if setStateDtoInDB == nil {
					// create
					if _, err = hu.SetStateServiceImpl.CreateSetState(setStateDto); err != nil {
						FeedbackCurrentClient(message, currentClient, ERROR_CREATE_STATE_FAILED)
						return err
					}
				} else {
					// update
					setStateDtoInDB.ConstructByValue(setStateDto.Value)
					if _, err = hu.SetStateServiceImpl.UpdateSetState(setStateDtoInDB); err != nil {
						FeedbackCurrentClient(message, currentClient, ERROR_UPDATE_STATE_FAILED)
						return err
					}
				}
			}
		}
	case TARGET_APPS:
		// serve on HTTP API, this signal only for broadcast
	case TARGET_RESOURCE:
		// serve on HTTP API, this signal only for broadcast
	}

	// feedback currentClient
	FeedbackCurrentClient(message, currentClient, ERROR_CREATE_OR_UPDATE_STATE_OK)

	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)
	return nil
}
