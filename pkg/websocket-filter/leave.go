package filter

import (
	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalLeave(hub *websocket.Hub, message *websocket.Message) error {
	currentClient := hub.Clients[message.ClientID]
	websocket.KickClient(hub, currentClient)
	return nil
}
