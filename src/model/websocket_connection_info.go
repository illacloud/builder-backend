package model

import (
	"fmt"

	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const PROTOCOL_WEBSOCKET = "ws"
const PROTOCOL_WEBSOCKET_OVER_TLS = "wss"
const DASHBOARD_WS_URL = "%s://%s/teams/%s/room/websocketConnection/dashboard"
const ROOM_WS_URL = "%s://%s/teams/%s/room/websocketConnection/apps/%s"
const ROOM_BINARY_WS_URL = "%s://%s/teams/%s/room/binaryWebsocketConnection/apps/%s"
const SELF_HOST_DASHBOARD_WS_URL = "/builder-ws/teams/%s/room/websocketConnection/dashboard"
const SELF_HOST_ROOM_WS_URL = "/builder-ws/teams/%s/room/websocketConnection/apps/%s"
const SELF_HOST_ROOM_BINARY_WS_URL = "/builder-ws/teams/%s/room/binaryWebsocketConnection/apps/%s"

type WebsocketConnectionInfo struct {
	Config *config.Config
}

func NewWebsocketConnectionInfo(conf *config.Config) *WebsocketConnectionInfo {
	return &WebsocketConnectionInfo{
		Config: conf,
	}
}

func (i *WebsocketConnectionInfo) GetDashboardConnectionAddress(teamID int) string {
	if i.Config.IsSelfHostMode() {
		return fmt.Sprintf(SELF_HOST_DASHBOARD_WS_URL, idconvertor.ConvertIntToString(teamID))
	} else {
		return fmt.Sprintf(DASHBOARD_WS_URL, i.Config.GetWebsocketProtocol(), i.Config.GetWebScoketServerListenAddress(), idconvertor.ConvertIntToString(teamID))
	}
}

func (i *WebsocketConnectionInfo) GetAppRoomConnectionAddress(teamID int, roomID int) string {
	if i.Config.IsSelfHostMode() {
		return fmt.Sprintf(SELF_HOST_ROOM_WS_URL, idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	} else {
		return fmt.Sprintf(ROOM_WS_URL, i.Config.GetWebsocketProtocol(), i.Config.GetWebScoketServerListenAddress(), idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	}
}

func (i *WebsocketConnectionInfo) GetAppRoomBinaryConnectionAddress(teamID int, roomID int) string {
	if i.Config.IsSelfHostMode() {
		return fmt.Sprintf(SELF_HOST_ROOM_BINARY_WS_URL, idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	} else {
		return fmt.Sprintf(ROOM_BINARY_WS_URL, i.Config.GetWebsocketProtocol(), i.Config.GetWebScoketServerListenAddress(), idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	}
}
