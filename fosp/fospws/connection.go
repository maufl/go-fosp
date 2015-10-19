// Copyright (C) 2015 Felix Maurer
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

package fospws

import (
	"bytes"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/maufl/go-fosp/fosp"
	"github.com/op/go-logging"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ErrChanError is returned when a response should be read from a channel but the channel returns an error instead.
var ErrChanError = errors.New("channel error when recieving response")

// ErrRequestTimeout is returned when a response should be read from a channel but the timeout is reached.
var ErrRequestTimeout = errors.New("request timed out")

// MessageHandler is the interface of objects that know how to process Messages.
type MessageHandler interface {
	HandleMessage(*NumberedMessage)
}

type NumberedMessage struct {
	fosp.Message
	Seq        uint64
	BinaryBody bool
}

var connLog = logging.MustGetLogger("fospws/connection")

// Connection represents a generic FOSP connection.
// It is the base for ServerConnection and for Client.
type Connection struct {
	ws *websocket.Conn

	currentSeq          uint64
	pendingRequests     map[uint64]chan *fosp.Response
	pendingRequestsLock sync.RWMutex

	out            chan *NumberedMessage
	messageHandler MessageHandler

	RequestTimeout time.Duration
}

// NewConnection creates a new FOSP connection from an existing WebSocket connection.
func NewConnection(ws *websocket.Conn) *Connection {
	if ws == nil {
		panic("Cannot initialize fosp connection without websocket")
	}
	con := &Connection{ws: ws, pendingRequests: make(map[uint64]chan *fosp.Response), out: make(chan *NumberedMessage), RequestTimeout: time.Second * 15}
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

func (c *Connection) panicRecover() {
	if r := recover(); r != nil {
		buf := make([]byte, 1024)
		runtime.Stack(buf, false)
		connLog.Critical("Panic in listening goroutine: %s\n%s", r, string(buf))
		c.Close()
	}
}

func (c *Connection) listen() {
	defer c.panicRecover()
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			connLog.Critical("Error while receiving new WebSocket message :: %s", err.Error())
			c.Close()
			break
		}
		reader := bytes.NewBuffer(message)
		if msg, seq, err := parseMessage(reader); err != nil {
			connLog.Error("Error while parsing message :: %s", err.Error())
			c.Close()
			break
		} else {
			connLog.Debug("Received new message")
			c.handleResponse(msg, seq)
			if c.messageHandler != nil {
				go c.messageHandler.HandleMessage(&NumberedMessage{Message: msg, Seq: seq})
			} else {
				connLog.Warning("No message handler registered")
			}
		}
	}
}

func (c *Connection) talk() {
	defer c.panicRecover()
	for {
		if oMsg, ok := <-c.out; ok {
			if request, ok := oMsg.Message.(*fosp.Request); ok && request.Method == fosp.WRITE {
				c.ws.WriteMessage(websocket.BinaryMessage, serializeMessage(request, oMsg.Seq))
			} else if oMsg.BinaryBody {
				c.ws.WriteMessage(websocket.BinaryMessage, serializeMessage(oMsg.Message, oMsg.Seq))
			} else {
				c.ws.WriteMessage(websocket.TextMessage, serializeMessage(oMsg.Message, oMsg.Seq))
			}
		} else {
			connLog.Critical("Output channel of connection broken!")
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
func (c *Connection) Send(msg fosp.Message, seq ...uint64) {
	oMsg := &NumberedMessage{Message: msg}
	if len(seq) > 0 {
		oMsg.Seq = seq[0]
	}
	c.out <- oMsg
}

func (c *Connection) SendNumberedMessage(msg *NumberedMessage) {
	c.out <- msg
}

// SendRequest will send a Request and block until a Response is returned or timedout.
func (c *Connection) SendRequest(req *fosp.Request) (*fosp.Response, error) {
	seq := atomic.AddUint64(&c.currentSeq, uint64(1))

	c.pendingRequestsLock.Lock()
	c.pendingRequests[seq] = make(chan *fosp.Response)
	c.pendingRequestsLock.Unlock()
	connLog.Info("Sending request: %s", req)
	c.Send(req, seq)
	var (
		resp    *fosp.Response
		ok      = false
		timeout = false
	)
	c.pendingRequestsLock.RLock()
	returnChan := c.pendingRequests[seq]
	c.pendingRequestsLock.RUnlock()
	select {
	case resp, ok = <-returnChan:
	case <-time.After(c.RequestTimeout):
		timeout = true
	}
	connLog.Debug("Received response or timeout")

	c.pendingRequestsLock.Lock()
	delete(c.pendingRequests, seq)
	c.pendingRequestsLock.Unlock()

	if !ok {
		connLog.Error("Something went wrong when reading channel")
		return nil, ErrChanError
	}
	if timeout {
		connLog.Warning("Request timed out")
		return nil, ErrRequestTimeout
	}
	connLog.Info("Recieved response: %s", resp)
	return resp, nil
}

func (c *Connection) handleResponse(msg fosp.Message, seq uint64) {
	if resp, ok := msg.(*fosp.Response); ok {
		connLog.Info("Received new response: %s", resp)
		c.pendingRequestsLock.RLock()
		if ch, ok := c.pendingRequests[uint64(seq)]; ok {
			connLog.Debug("Returning response to caller")
			ch <- resp
		}
		c.pendingRequestsLock.RUnlock()
	}
}
