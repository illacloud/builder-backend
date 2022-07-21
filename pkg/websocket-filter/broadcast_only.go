package filter

import (
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalBroadcastOnly(hub *ws.Hub, message *ws.Message) error {
	// deserialize message
	currentClient := hub.Clients[message.ClientID]
	message.RewriteBroadcast()

	// feedback otherClient
	hub.BroadcastToOtherClients(message, currentClient)
	return nil
}
