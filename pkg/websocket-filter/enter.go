package filter

import (
	"errors"

	ws "github.com/illa-family/builder-backend/internal/websocket"
)

func SignalEnter(hub *ws.Hub, message *ws.Message) error {
	// init
	currentClient := hub.Clients[message.ClientID]
	var ok bool
	if len(message.Payload) == 0 {
		err := errors.New("[websocket-server] websocket protocol syntax error.")
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, err)
		return err
	}
	var authToken map[string]interface{}
	if authToken, ok = message.Payload[0].(map[string]interface{}); !ok {
		err := errors.New("[websocket-server] websocket protocol syntax error.")
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, err)
		return err
	}
	token, _ := authToken["authToken"].(string)

	// convert authToken to uid
	userID, extractErr := user.ExtractUserIDFromToken(token)
	if extractErr != nil {
		return extractErr
	}
	validAccessToken, validaAccessErr := user.ValidateAccessToken(token)
	if validaAccessErr != nil {
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, validaAccessErr)
		return validaAccessErr
	}
	if !validAccessToken {
		err := errors.New("[websocket-server] access token invalied.")
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, err)
		return err
	}
	// assign logged in and mapped user id
	currentClient.IsLoggedIn = true
	currentClient.MappedUserID = userID
	currentClient.Feedback(message, ws.ERROR_CODE_LOGGEDIN)
	return nil

}
