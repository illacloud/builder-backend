package filter

import (
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

// @todo: the client should check userID, make sure do not broadcast to self.
func (hub *ws.Hub) Run() {
	for {
		select {
		// handle register event
		case client := <-hub.Register:
			hub.Clients[client.ID] = client
		// handle unregister events
		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client.ID]; ok {
				delete(hub.Clients, client.ID)
				close(client.Send)
			}
		// handle all hub broadcast events
		case message := <-hub.Broadcast:
			for _, client := range hub.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(hub.Clients, client.ID)
				}
			}
		// handle client on message event
		case message := <-hub.OnMessage:
			filter.SignalFilter(hub, message)
		}

	}
}

func SignalFilter(hub *ws.Hub, message *ws.Message) error {
	switch message.Signal {
	case SIGNAL_PING:
		return filter.SignalPing(hub, message)
	case SIGNAL_ENTER:
		return filter.SignalEnter(hub, message)
	case SIGNAL_LEAVE:
		return filter.SignalLeave(hub, message)
	case SIGNAL_CREATE_STATE:
		return filter.SignalCreateState(hub, message)
	case SIGNAL_DELETE_STATE:
		return filter.SignalDeleteState(hub, message)
	case SIGNAL_UPDATE_STATE:
		return filter.SignalUpdateState(hub, message)
	case SIGNAL_MOVE_STATE:
		return filter.SignalMoveState(hub, message)
	case SIGNAL_CREATE_OR_UPDATE:
		return filter.SignalCreateOrUpdate(hub, message)
	case SIGNAL_ONLY_BROADCAST:
		return filter.SignalOnlyBroadcast(hub, message)
	default:
		return nil

	}
	return nil
}

func OptionFilter(hub *ws.Hub, client *ws.Client, message *ws.Message) error {
	return nil
}

func KickClient(hub *ws.Hub, client *ws.Client) {
	close(client.Send)
	delete(hub.Clients, client.ID)
}
