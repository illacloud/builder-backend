package filter

import (
	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalPing(hub *ws.Hub, message *ws.Message) error {
	currentClient.Feedback(message, ERROR_CODE_PONG, nil)
	return nil
}
