package filter

import (
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalLeave(hub *ws.Hub, message *ws.Message) error {
	currentClient := hub.Clients[message.ClientID]
	ws.KickClient(hub, currentClient)
	return nil
}
