package filter

import (
	"errors"

	"github.com/illa-family/builder-backend/internal/websocket"
)

func SignalEnter(hub *websocket.Hub, message *websocket.Message) error {
	// init
	currentClient := hub.Clients[message.ClientID]
	var ok bool
	if len(message.Payload) == 0 {
		errorMessage := errors.New("[websocket-server] websocket protocol syntax error.")
		websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_CODE_LOGIN_FAILED, errorMessage)
		return errorMessage
	}
	var authToken map[string]interface{}
	if authToken, ok = message.Payload[0].(map[string]interface{}); !ok {
		errorMessage := errors.New("[websocket-server] websocket protocol syntax error.")
		websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_CODE_LOGIN_FAILED, errorMessage)
		return errorMessage
	}
	token, _ := authToken["authToken"].(string)

	// convert authToken to uid
	userID, extractErr := user.ExtractUserIDFromToken(token)
	if extractErr != nil {
		return extractErr
	}
	validAccessToken, validaAccessErr := user.ValidateAccessToken(token)
	if validaAccessErr != nil {
		websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_CODE_LOGIN_FAILED, validaAccessErr)
		return validaAccessErr
	}
	if !validAccessToken {
		errorMessage := errors.New("[websocket-server] access token invalied.")
		websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_CODE_LOGIN_FAILED, errorMessage)
		return errorMessage
	}
	// assign logged in and mapped user id
	currentClient.IsLoggedIn = true
	currentClient.MappedUserID = userID
	websocket.FeedbackCurrentClient(message, currentClient, websocket.ERROR_CODE_LOGGEDIN)
	return nil

}
