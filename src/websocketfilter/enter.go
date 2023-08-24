// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filter

import (
	"errors"

	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/internal/util/supervisior"
	ws "github.com/illacloud/builder-backend/internal/websocket"
	"github.com/illacloud/builder-backend/pkg/user"
)

func SignalEnter(hub *ws.Hub, message *ws.Message) error {
	// init
	currentClient, hit := hub.Clients[message.ClientID]
	if !hit {
		return errors.New("[SignalEnter] target client(" + message.ClientID.String() + ") does dot exists.")
	}
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

	// init supervisior client
	sv, err := supervisior.NewSupervisior()
	if err != nil {
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, err)
		return err
	}
	// validate user token
	validated, errInValidate := sv.ValidateUserAccount(token)
	if errInValidate != nil {
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, errInValidate)
		return errInValidate
	}
	if !validated {
		err := errors.New("access token invalid.")
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, err)
		return err
	}

	// extract userID
	userID, userUID, errInExtract := user.ExtractUserIDFromToken(token)
	if errInExtract != nil {
		err := errors.New("access token extract failed.")
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, err)
		return err
	}

	// fetch user remote data
	user, errInGetUserInfo := datacontrol.GetUserInfo(userID)
	if errInGetUserInfo != nil {
		currentClient.Feedback(message, ws.ERROR_CODE_LOGIN_FAILED, errInGetUserInfo)
		return errInGetUserInfo
	}

	// assign logged in and mapped user id
	currentClient.IsLoggedIn = true
	currentClient.MappedUserID = userID
	currentClient.MappedUserUID = userUID

	// storage user to app edited by lists
	app, errInRetrieveApp := hub.AppRepositoryImpl.RetrieveAppByIDAndTeamID(currentClient.APPID, currentClient.TeamID)
	if errInRetrieveApp == nil {
		appEditedBy := repository.NewAppEditedByUserID(userID)
		app.PushEditedBy(appEditedBy)
		hub.AppRepositoryImpl.UpdateWholeApp(app)
	}

	// broadcast in room users
	inRoomUsers := hub.GetInRoomUsersByRoomID(currentClient.APPID)
	inRoomUsers.EnterRoom(user)
	message.SetBroadcastPayload(inRoomUsers.FetchAllInRoomUsers())
	message.RewriteBroadcast()
	hub.BroadcastToRoomAllClients(message, currentClient)

	// broadcast attached components users
	message.SetBroadcastType(ws.BROADCAST_TYPE_ATTACH_COMPONENT)
	message.SetBroadcastPayload(inRoomUsers.FetchAllAttachedUsers())
	message.RewriteBroadcast()
	hub.BroadcastToRoomAllClients(message, currentClient)
	return nil

}
