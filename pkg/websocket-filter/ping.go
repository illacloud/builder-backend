package filter

import (
	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalPing(hub *websocket.Hub, message *websocket.Message) error {
	websocket.FeedbackCurrentClient(message, currentClient, ERROR_CODE_PONG, nil)
	return nil
}
