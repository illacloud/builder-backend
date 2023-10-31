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

package websocket

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const (
	CLIENT_TYPE_TEXT   = 1
	CLIENT_TYPE_BINARY = 2
)

const (
	CLIENT_STATUS_ACTIVE = 1
	CLIENT_STATUS_DEAD   = 2
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 60 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10485760 // 10 MiB
)

const DEFAULT_INSTANCE_ID = "SELF_HOST"
const DEFAULT_APP_ID = 0
const DASHBOARD_APP_ID = -1

var (
	newLine       = []byte{'\n'}
	binaryNewLine = []byte{'\x1e', '\x1e', '\x1e', '\x1e'}
	charSpace     = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10485760,
	WriteBufferSize: 10485760,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ID uuid.UUID

	Status int

	Type int

	MappedUserID int

	MappedUserUID uuid.UUID

	IsLoggedIn bool

	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// teamID, 0 by default in SELF_HOST mode
	TeamID int // TeamID

	// appID, 0 by default, -1 for dashboard
	APPID int
}

func (c *Client) GetID() uuid.UUID {
	return c.ID
}

func (c *Client) GetAPPID() int {
	return c.APPID
}

func (c *Client) SetType(clientType int) {
	c.Type = clientType
}

func (c *Client) SetStatus(status int) {
	c.Status = status
}

func (c *Client) IsDead() bool {
	return c.Status == CLIENT_STATUS_DEAD
}

func (c *Client) ExportMappedUserIDToString() string {
	return idconvertor.ConvertIntToString(c.MappedUserID)
}

func NewClient(hub *Hub, conn *websocket.Conn, teamID int, appID int, clientType int) *Client {
	return &Client{
		ID:           uuid.New(),
		Status:       CLIENT_STATUS_ACTIVE,
		Type:         clientType,
		MappedUserID: 0,
		IsLoggedIn:   false,
		Hub:          hub,
		Conn:         conn,
		Send:         make(chan []byte, 10240),
		TeamID:       teamID,
		APPID:        appID}
}

func (c *Client) Feedback(message *Message, errorCode int, errorMessage error) {
	m := ""
	if errorMessage != nil {
		m = errorMessage.Error()
	}
	feedCurrentClient := Feedback{
		ErrorCode:    errorCode,
		ErrorMessage: m,
		Broadcast:    message.Broadcast,
		Data:         nil,
	}
	feedbyte, _ := feedCurrentClient.Serialization()
	c.Send <- feedbyte
}

func (c *Client) FeedbackBinary(message []byte) {
	c.Send <- message
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// got message
		messageType, message, errInReadMessage := c.Conn.ReadMessage()
		if errInReadMessage != nil {
			if websocket.IsUnexpectedCloseError(errInReadMessage, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ReadMessage] error: %v", errInReadMessage)
			}
			break
		}
		// check out message type
		switch messageType {
		case websocket.TextMessage:
			c.OnTextMessage(message)
		case websocket.BinaryMessage:
			c.OnBinaryMessage(message)
		}
	}
}

func (c *Client) OnTextMessage(message []byte) {
	message = bytes.TrimSpace(bytes.Replace(message, newLine, charSpace, -1))
	msg, _ := NewMessage(c.ID, c.APPID, message)
	// send to hub and process
	if msg != nil {
		c.Hub.OnTextMessage <- msg
	}
}

func (c *Client) OnBinaryMessage(message []byte) {
	// unpack binary message and fill clientID
	binaryMessageType, errInGetType := GetBinaryMessageType(message)

	if errInGetType != nil {
		log.Printf("[OnBinaryMessage] error: %v", errInGetType)
		return
	}

	// fill client ID
	switch binaryMessageType {
	case BINARY_MESSAGE_TYPE_MOVING:
		// decode binary message
		movingMessageBin := &MovingMessageBin{}
		if errInUnmarshal := proto.Unmarshal(message, movingMessageBin); errInUnmarshal != nil {
			log.Printf("[OnBinaryMessage] Failed to parse message MovingMessageBin: ", errInUnmarshal)
			return
		}
		movingMessageBin.ClientID = c.ID.String()

		// encode binary message
		var errInMarshal error
		message, errInMarshal = proto.Marshal(movingMessageBin)
		if errInMarshal != nil {
			log.Printf("[OnBinaryMessage] Failed to parse message MovingMessageBin: ", errInMarshal)
			return
		}
	}

	// send to following pipeline
	c.Hub.OnBinaryMessage <- message
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
//
// All message types (TextMessage, BinaryMessage, CloseMessage, PingMessage and
// PongMessage) are supported.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.SetStatus(CLIENT_STATUS_DEAD)
				return
			}
			messageType := checkOutMessageType(message)

			w, errInSetNextWriter := c.Conn.NextWriter(messageType)
			if errInSetNextWriter != nil {
				c.SetStatus(CLIENT_STATUS_DEAD)
				log.Printf("[WritePump] c.Conn.NextWriter error: %v\n", errInSetNextWriter)
				return
			}
			_, errInWrite := w.Write(message)
			if errInWrite != nil {
				c.SetStatus(CLIENT_STATUS_DEAD)
				log.Printf("[WritePump] Write message error: %v\n", errInWrite)
				return
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				if messageType == websocket.TextMessage {
					w.Write(newLine)
				} else {
					w.Write(binaryNewLine)
				}
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				c.SetStatus(CLIENT_STATUS_DEAD)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if errInWritePingMessage := c.Conn.WriteMessage(websocket.PingMessage, nil); errInWritePingMessage != nil {
				c.SetStatus(CLIENT_STATUS_DEAD)
				log.Printf("[WritePump] c.Conn.WriteMessage(websocket.PingMessage) failed: %v\n", errInWritePingMessage)
				return
			}
		}
	}
}

func ping(ws *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Printf("[ping] WriteControl ping error: %v\n", err)
			}
		case <-done:
			return
		}
	}
}

func checkOutMessageType(message []byte) int {
	untypedData := make(map[string]interface{})
	errInUnmarshal := json.Unmarshal(message, &untypedData)
	if errInUnmarshal != nil {
		return websocket.BinaryMessage
	}
	return websocket.TextMessage
}
