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

	"github.com/illacloud/builder-backend/src/websocket"
)

func SignalPing(hub *websocket.Hub, message *websocket.Message) error {
	currentClient, hit := hub.Clients[message.ClientID]
	if !hit {
		return errors.New("[SignalPing] target client(" + message.ClientID.String() + ") does dot exists.")
	}

	currentClient.Feedback(message, websocket.ERROR_CODE_PONG, nil)
	return nil
}
