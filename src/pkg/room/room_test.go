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

package room

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSample(t *testing.T) {
	assert.Nil(t, nil)
}

func TestGetProtocol(t *testing.T) {
	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "")
	os.Setenv("WEBSOCKET_PORT", "")
	os.Setenv("WSS_ENABLED", "")

	protocol0 := getProtocol()
	assert.Equal(t, "ws", protocol0, "the protcol should be ws")

	os.Setenv("WSS_ENABLED", "true")
	protocol1 := getProtocol()
	assert.Equal(t, "wss", protocol1, "the protcol should be wss")

	os.Setenv("WSS_ENABLED", "false")
	protocol2 := getProtocol()
	assert.Equal(t, "ws", protocol2, "the protcol should be ws")

}

func TestGetServerAddress(t *testing.T) {
	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "")
	os.Setenv("WEBSOCKET_PORT", "")

	addr0 := getServerAddress()
	assert.Equal(t, "localhost", addr0, "the server address should be localhost")

	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "127.0.0.1")
	addr1 := getServerAddress()
	assert.Equal(t, "127.0.0.1", addr1, "the server address should be 127.0.0.1")

	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "thisismyhost.com")
	addr2 := getServerAddress()
	assert.Equal(t, "thisismyhost.com", addr2, "the server address should be thisismyhost.com")

}

func TestGetWebSocketPort(t *testing.T) {
	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "")
	os.Setenv("WEBSOCKET_PORT", "")

	port0 := getWebSocketPort()
	assert.Equal(t, "8000", port0, "the default websocket port should be 8000")

	os.Setenv("WEBSOCKET_PORT", "8000")
	port1 := getWebSocketPort()
	assert.Equal(t, "8000", port1, "the websocket port should be 8000")

	os.Setenv("WEBSOCKET_PORT", "9999")
	port2 := getWebSocketPort()
	assert.Equal(t, "9999", port2, "the websocket port should be 9999")

}

func TestGetDashboardConn(t *testing.T) {
	// set env to default
	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "")
	os.Setenv("WEBSOCKET_PORT", "")

	rsi := RoomServiceImpl{}
	r, err := rsi.GetDashboardConn("SELF_SERVE")
	assert.Nil(t, err)
	assert.Equal(t, "ws://localhost:8000/room/SELF_SERVE/dashboard", r.WSURL, "the server address should be ws://localhost:8000/room/SELF_SERVE/dashboard")

	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "myhostname.com")
	os.Setenv("WEBSOCKET_PORT", "443")

	r2, err2 := rsi.GetDashboardConn("XHRT12A")
	assert.Nil(t, err2)
	assert.Equal(t, "ws://myhostname.com:443/room/XHRT12A/dashboard", r2.WSURL, "the server address should be ws://myhostname.com:443/room/XHRT12A/dashboard")

}

func TestGetAppRoomConn(t *testing.T) {
	// set env to default
	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "")
	os.Setenv("WEBSOCKET_PORT", "")

	rsi := RoomServiceImpl{}
	r, err := rsi.GetAppRoomConn("SELF_SERVE", 74)
	assert.Nil(t, err)
	assert.Equal(t, "ws://localhost:8000/room/SELF_SERVE/app/74", r.WSURL, "the server address should be ws://localhost:8000/room/SELF_SERVE/app/74")

	os.Setenv("WEBSOCKET_SERVER_ADDRESS", "myhostname2.com")
	os.Setenv("WEBSOCKET_PORT", "4430")

	r2, err2 := rsi.GetAppRoomConn("SELF_SERVE", 74)
	assert.Nil(t, err2)
	assert.Equal(t, "ws://myhostname2.com:4430/room/SELF_SERVE/app/74", r2.WSURL, "the server address should be ws://myhostname2.com:4430/room/SELF_SERVE/app/74")

}
