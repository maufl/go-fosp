// Copyright (C) 2014 Felix Maurer
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package fosp

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// MessageHandler is the interface of objects that know how to process Messages.
type MessageHandler interface {
	HandleMessage(Message)
}

// Connection represents a generic FOSP connection.
// It is the base for ServerConnection and for Client.
type Connection struct {
	ws *websocket.Conn

	negotiated    bool
	authenticated bool

	currentSeq          uint64
	pendingRequests     map[uint64]chan *Response
	pendingRequestsLock sync.Mutex

	out            chan Message
	messageHandler MessageHandler

	lg *logging.Logger
}

// NewConnection creates a new FOSP connection from an existing WebSocket connection.
func NewConnection(ws *websocket.Conn) *Connection {
	if ws == nil {
		panic("Cannot initialize fosp connection without websocket")
	}
	con := &Connection{ws: ws, pendingRequests: make(map[uint64]chan *Response), out: make(chan Message)}
	con.lg = logging.MustGetLogger("go-fosp/fosp/connection")
	logging.SetLevel(logging.NOTICE, "go-fosp/fosp/connection")
	con.messageHandler = con
	go con.listen()
	go con.talk()
	return con
}

// OpenConnection creates a new FOSP connection to the remoteDomain.
// It will open a WebSocket connection to this remoteDomain or return an error.
func OpenConnection(remoteDomain string) (*Connection, error) {
	url := "ws://" + remoteDomain + ":1337"
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		return nil, err
	}
	connection := NewConnection(ws)
	return connection, nil
}

// RegisterMessageHandler accepts a function that should be called when a Message is received.
func (c *Connection) RegisterMessageHandler(handler MessageHandler) {
	c.messageHandler = handler
}

func (c *Connection) listen() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			c.lg.Critical("Error while receiving new WebSocket message :: ", err.Error())
			c.Close()
			break
		}
		if msg, err := parseMessage(message); err != nil {
			c.lg.Error("Error while parsing message :: ", err.Error())
			c.Close()
			break
		} else {
			c.lg.Debug("Received new message")
			if c.messageHandler != nil {
				c.messageHandler.HandleMessage(msg)
			} else {
				c.lg.Warning("No message handler registered")
			}
		}
	}
}

func (c *Connection) talk() {
	for {
		if msg, ok := <-c.out; ok {
			if msg.Type() == Text {
				c.ws.WriteMessage(websocket.TextMessage, msg.Bytes())
			} else {
				c.ws.WriteMessage(websocket.BinaryMessage, msg.Bytes())
			}
		} else {
			c.lg.Critical("Output channel of connection broken!")
			c.Close()
			break
		}
	}
}

// Close this connection and clean up.
// TODO: Websocket should send close message before tearing down the connection
func (c *Connection) Close() {
	c.ws.Close()
}

// Send queues an Message to be send.
func (c *Connection) Send(msg Message) {
	c.out <- msg
}

// SendRequest will send a Request and block until a Response is returned or timedout.
func (c *Connection) SendRequest(rt RequestType, url *URL, headers map[string]string, body []byte) (*Response, error) {
	seq := atomic.AddUint64(&c.currentSeq, uint64(1))
	req := NewRequest(rt, url, int(seq), headers, body)

	c.pendingRequestsLock.Lock()
	c.pendingRequests[seq] = make(chan *Response)
	c.pendingRequestsLock.Unlock()
	c.lg.Info("Sending request: %s %s %d", req.request, req.url, req.seq)
	c.Send(req)
	var (
		resp    *Response
		ok      = false
		timeout = false
	)
	select {
	case resp, ok = <-c.pendingRequests[seq]:
	case <-time.After(time.Second * 15):
		timeout = true
	}
	c.lg.Debug("Received response or timeout")

	c.pendingRequestsLock.Lock()
	delete(c.pendingRequests, seq)
	c.pendingRequestsLock.Unlock()

	if !ok {
		c.lg.Error("Something went wrong when reading channel")
		return nil, errors.New("error when receiving response")
	}
	if timeout {
		c.lg.Warning("Request timed out")
		return nil, errors.New("request timed out")
	}
	c.lg.Info("Recieved response: %s %d %d", resp.response, resp.status, resp.seq)
	return resp, nil
}

// HandleMessage is a generic Message handler that does nothing except returning responses.
func (c *Connection) HandleMessage(msg Message) {
	if resp, ok := msg.(*Response); ok {
		c.lg.Info("Received new response: %s %d %d", resp.response, resp.status, resp.seq)
		if ch, ok := c.pendingRequests[uint64(resp.seq)]; ok {
			c.lg.Debug("Returning response to caller")
			ch <- resp
		}
	}
}
