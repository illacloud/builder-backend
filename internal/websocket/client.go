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

package ws

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1048576 // 1 MiB
)

const DEFAULT_INSTANCE_ID = "SELF_HOST"
const DEFAULT_APP_ID = 0
const DASHBOARD_APP_ID = -1

var (
	newline   = []byte{'\n'}
	charSpace = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1048576,
	WriteBufferSize: 1048576,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ID uuid.UUID

	MappedUserID int

	IsLoggedIn bool

	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// teamID, SELF_HOST by default
	TeamID int // TeamID

	// appID, 0 by default, -1 for dashboard
	APPID int
}

func (c *Client) GetAPPID() int {
	return c.APPID
}

func NewClient(hub *Hub, conn *websocket.Conn, teamID int, appID int) *Client {
	return &Client{
		ID:           uuid.Must(uuid.NewV4(), nil),
		MappedUserID: 0,
		IsLoggedIn:   false,
		Hub:          hub,
		Conn:         conn,
		Send:         make(chan []byte, 1024),
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
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// on message, format
		message = bytes.TrimSpace(bytes.Replace(message, newline, charSpace, -1))
		msg, _ := NewMessage(c.ID, c.APPID, message)
		// send to hub and process
		if msg != nil {
			c.Hub.OnMessage <- msg
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
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
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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
				log.Println("ping:", err)
			}
		case <-done:
			return
		}
	}
}
