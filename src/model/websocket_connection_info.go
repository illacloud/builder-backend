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

const (
	CONNECTION_ZONE_NORTH_ASIA = "northAsia"
	CONNECTION_ZONE_EAST_ASIA  = "eastAsia"
)

var CountryCodeMappedIngressAddress = map[string]string{
	"IN": CONNECTION_ZONE_NORTH_ASIA,
	"CN": CONNECTION_ZONE_EAST_ASIA,
	"JP": CONNECTION_ZONE_EAST_ASIA,
}

type WebsocketConnectionInfo struct {
	Config                     *config.Config
	DefaultConnectionAddress   string
	NorthAsiaConnectionAddress string
	EastAsiaConnectionAddress  string
}

func NewWebsocketConnectionInfo(conf *config.Config) *WebsocketConnectionInfo {
	return &WebsocketConnectionInfo{
		Config:                     conf,
		DefaultConnectionAddress:   conf.GetWebScoketServerConnectionAddress(),
		NorthAsiaConnectionAddress: conf.GetWebScoketServerConnectionAddressNorthAsia(),
		EastAsiaConnectionAddress:  conf.GetWebScoketServerConnectionAddressEastAsia(),
	}
}

func (i *WebsocketConnectionInfo) GetConnectionAddressByCountryCode(countryCode string) string {
	// check user ip zone
	connectionZone, hitConnectionZone := CountryCodeMappedIngressAddress[countryCode]
	if !hitConnectionZone {
		return i.DefaultConnectionAddress
	}
	// get address list
	switch connectionZone {
	case CONNECTION_ZONE_NORTH_ASIA:
		return i.NorthAsiaConnectionAddress
	case CONNECTION_ZONE_EAST_ASIA:
		return i.EastAsiaConnectionAddress
	default:
		return i.DefaultConnectionAddress
	}
}

func (i *WebsocketConnectionInfo) GetDashboardConnectionAddress(teamID int, zone string) string {
	if i.Config.IsSelfHostMode() {
		return fmt.Sprintf(SELF_HOST_DASHBOARD_WS_URL, idconvertor.ConvertIntToString(teamID))
	} else {
		return fmt.Sprintf(DASHBOARD_WS_URL, i.Config.GetWebsocketProtocol(), i.GetConnectionAddressByCountryCode(zone), idconvertor.ConvertIntToString(teamID))
	}
}

func (i *WebsocketConnectionInfo) GetAppRoomConnectionAddress(teamID int, roomID int, zone string) string {
	if i.Config.IsSelfHostMode() {
		return fmt.Sprintf(SELF_HOST_ROOM_WS_URL, idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	} else {
		return fmt.Sprintf(ROOM_WS_URL, i.Config.GetWebsocketProtocol(), i.GetConnectionAddressByCountryCode(zone), idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	}
}

func (i *WebsocketConnectionInfo) GetAppRoomBinaryConnectionAddress(teamID int, roomID int, zone string) string {
	if i.Config.IsSelfHostMode() {
		return fmt.Sprintf(SELF_HOST_ROOM_BINARY_WS_URL, idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	} else {
		return fmt.Sprintf(ROOM_BINARY_WS_URL, i.Config.GetWebsocketProtocol(), i.GetConnectionAddressByCountryCode(zone), idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	}
}
