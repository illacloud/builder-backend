package filter

import (
	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalBroadcastOnly(hub *websocket.Hub, message *websocket.Message) error {
	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	message.RewriteBroadcast()

	// feedback otherClient
	BroadcastToOtherClients(hub, message, currentClient)
	return nil
}
