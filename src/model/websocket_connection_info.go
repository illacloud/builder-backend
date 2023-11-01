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
	CONNECTION_ZONE_SOUTH_ASIA    = "southAsia"
	CONNECTION_ZONE_EAST_ASIA     = "eastAsia"
	CONNECTION_ZONE_CENTER_EUROPE = "centerEurope"
)

var CountryCodeMappedIngressAddress = map[string]string{
	// east asia
	"AF": CONNECTION_ZONE_EAST_ASIA,
	"CN": CONNECTION_ZONE_EAST_ASIA,
	"JP": CONNECTION_ZONE_EAST_ASIA,
	"KP": CONNECTION_ZONE_EAST_ASIA,
	"HK": CONNECTION_ZONE_EAST_ASIA,
	"MO": CONNECTION_ZONE_EAST_ASIA,
	"MN": CONNECTION_ZONE_EAST_ASIA,
	"KR": CONNECTION_ZONE_EAST_ASIA,
	"TW": CONNECTION_ZONE_EAST_ASIA,
	// south asia
	"TL": CONNECTION_ZONE_SOUTH_ASIA,
	"ID": CONNECTION_ZONE_SOUTH_ASIA,
	"LA": CONNECTION_ZONE_SOUTH_ASIA,
	"IN": CONNECTION_ZONE_SOUTH_ASIA,
	"MY": CONNECTION_ZONE_SOUTH_ASIA,
	"MM": CONNECTION_ZONE_SOUTH_ASIA,
	"PH": CONNECTION_ZONE_SOUTH_ASIA,
	"SG": CONNECTION_ZONE_SOUTH_ASIA,
	"BD": CONNECTION_ZONE_SOUTH_ASIA,
	"BT": CONNECTION_ZONE_SOUTH_ASIA,
	"IR": CONNECTION_ZONE_SOUTH_ASIA,
	"MV": CONNECTION_ZONE_SOUTH_ASIA,
	"NP": CONNECTION_ZONE_SOUTH_ASIA,
	"PK": CONNECTION_ZONE_SOUTH_ASIA,
	"LK": CONNECTION_ZONE_SOUTH_ASIA,
	"KZ": CONNECTION_ZONE_SOUTH_ASIA,
	"KG": CONNECTION_ZONE_SOUTH_ASIA,
	"TJ": CONNECTION_ZONE_SOUTH_ASIA,
	"TM": CONNECTION_ZONE_SOUTH_ASIA,
	"UZ": CONNECTION_ZONE_SOUTH_ASIA,
	// center europe
	"AX": CONNECTION_ZONE_CENTER_EUROPE,
	"AL": CONNECTION_ZONE_CENTER_EUROPE,
	"AD": CONNECTION_ZONE_CENTER_EUROPE,
	"AT": CONNECTION_ZONE_CENTER_EUROPE,
	"BY": CONNECTION_ZONE_CENTER_EUROPE,
	"BE": CONNECTION_ZONE_CENTER_EUROPE,
	"BA": CONNECTION_ZONE_CENTER_EUROPE,
	"BG": CONNECTION_ZONE_CENTER_EUROPE,
	"HR": CONNECTION_ZONE_CENTER_EUROPE,
	"CV": CONNECTION_ZONE_CENTER_EUROPE,
	"CZ": CONNECTION_ZONE_CENTER_EUROPE,
	"DK": CONNECTION_ZONE_CENTER_EUROPE,
	"EE": CONNECTION_ZONE_CENTER_EUROPE,
	"FO": CONNECTION_ZONE_CENTER_EUROPE,
	"FI": CONNECTION_ZONE_CENTER_EUROPE,
	"FR": CONNECTION_ZONE_CENTER_EUROPE,
	"DE": CONNECTION_ZONE_CENTER_EUROPE,
	"GI": CONNECTION_ZONE_CENTER_EUROPE,
	"GR": CONNECTION_ZONE_CENTER_EUROPE,
	"GG": CONNECTION_ZONE_CENTER_EUROPE,
	"HU": CONNECTION_ZONE_CENTER_EUROPE,
	"IS": CONNECTION_ZONE_CENTER_EUROPE,
	"IE": CONNECTION_ZONE_CENTER_EUROPE,
	"IT": CONNECTION_ZONE_CENTER_EUROPE,
	"JE": CONNECTION_ZONE_CENTER_EUROPE,
	"XK": CONNECTION_ZONE_CENTER_EUROPE,
	"LV": CONNECTION_ZONE_CENTER_EUROPE,
	"LI": CONNECTION_ZONE_CENTER_EUROPE,
	"LT": CONNECTION_ZONE_CENTER_EUROPE,
	"LU": CONNECTION_ZONE_CENTER_EUROPE,
	"MT": CONNECTION_ZONE_CENTER_EUROPE,
	"IM": CONNECTION_ZONE_CENTER_EUROPE,
	"MD": CONNECTION_ZONE_CENTER_EUROPE,
	"MC": CONNECTION_ZONE_CENTER_EUROPE,
	"ME": CONNECTION_ZONE_CENTER_EUROPE,
	"NL": CONNECTION_ZONE_CENTER_EUROPE,
	"MK": CONNECTION_ZONE_CENTER_EUROPE,
	"NO": CONNECTION_ZONE_CENTER_EUROPE,
	"PL": CONNECTION_ZONE_CENTER_EUROPE,
	"PT": CONNECTION_ZONE_CENTER_EUROPE,
	"RO": CONNECTION_ZONE_CENTER_EUROPE,
	"RU": CONNECTION_ZONE_CENTER_EUROPE,
	"SM": CONNECTION_ZONE_CENTER_EUROPE,
	"RS": CONNECTION_ZONE_CENTER_EUROPE,
	"SK": CONNECTION_ZONE_CENTER_EUROPE,
	"SI": CONNECTION_ZONE_CENTER_EUROPE,
	"ES": CONNECTION_ZONE_CENTER_EUROPE,
	"SJ": CONNECTION_ZONE_CENTER_EUROPE,
	"SE": CONNECTION_ZONE_CENTER_EUROPE,
	"CH": CONNECTION_ZONE_CENTER_EUROPE,
}

type WebsocketConnectionInfo struct {
	Config                        *config.Config
	DefaultConnectionAddress      string
	SouthAsiaConnectionAddress    string
	EastAsiaConnectionAddress     string
	CenterEuropeConnectionAddress string
}

func NewWebsocketConnectionInfo(conf *config.Config) *WebsocketConnectionInfo {
	return &WebsocketConnectionInfo{
		Config:                        conf,
		DefaultConnectionAddress:      conf.GetWebScoketServerConnectionAddress(),
		SouthAsiaConnectionAddress:    conf.GetWebScoketServerConnectionAddressSouthAsia(),
		EastAsiaConnectionAddress:     conf.GetWebScoketServerConnectionAddressEastAsia(),
		CenterEuropeConnectionAddress: conf.GetWebScoketServerConnectionAddressCenterEurope(),
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
	case CONNECTION_ZONE_SOUTH_ASIA:
		return i.SouthAsiaConnectionAddress
	case CONNECTION_ZONE_EAST_ASIA:
		return i.EastAsiaConnectionAddress
	case CONNECTION_ZONE_CENTER_EUROPE:
		return i.CenterEuropeConnectionAddress
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
