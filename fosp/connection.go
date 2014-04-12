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
	_ "encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type connection struct {
	ws     *websocket.Conn
	server *server

	negotiated    bool
	authenticated bool
	user          string
	remote_domain string

	currentSeq          uint64
	pendingRequests     map[uint64]chan *Response
	pendingRequestsLock sync.Mutex

	out chan Message

	lg *logging.Logger
}

func NewConnection(ws *websocket.Conn, srv *server) *connection {
	if ws == nil || srv == nil {
		panic("Cannot initialize fosp connection without websocket or server")
	}
	con := new(connection)
	con.ws = ws
	con.server = srv
	con.currentSeq = 0
	con.pendingRequests = make(map[uint64]chan *Response)
	con.out = make(chan Message)
	con.lg = logging.MustGetLogger("go-fosp/fosp/connection")
	go con.listen()
	go con.talk()
	return con
}

func OpenConnection(srv *server, remote_domain string) (*connection, error) {
	url := "ws://" + remote_domain + ":1337"
	log.Println("Opening new connection to " + url)
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		log.Println("Error when opening new WebSocket connection " + err.Error())
		return nil, err
	}
	connection := NewConnection(ws, srv)
	connection.negotiated = true
	connection.authenticated = true
	connection.remote_domain = remote_domain
	resp, err := connection.SendRequest(Connect, &Url{}, map[string]string{}, []byte("{\"version\":\"0.1\"}"))
	if err != nil {
		return nil, errors.New("Error when negotiating connection")
	} else if resp.response != Succeeded {
		log.Println("Authentication failed!")
		return nil, errors.New("Authentication failed!")
	}
	log.Println("Successfully negotiated")
	resp, err = connection.SendRequest(Authenticate, &Url{}, map[string]string{}, []byte("{\"type\":\"server\", \"domain\":\""+srv.Domain()+"\"}"))
	if err != nil || resp.response != Succeeded {
		return nil, errors.New("Error when authenticating")
	}
	log.Println("Successfully authenticated")
	srv.registerConnection(connection, "@"+remote_domain)
	return connection, nil
}

func (c *connection) listen() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Println("Error while receiving new WebSocket message :: ", err)
			c.close()
			break
		}
		if msg, err := parseMessage(message); err != nil {
			log.Println("Error while parsing message :: ", err)
			c.close()
			break
		} else {
			c.lg.Debug("Received new message")
			c.checkState(msg)
		}
	}
}

func (c *connection) talk() {
	for {
		if msg, ok := <-c.out; ok {
			if msg.Type() == Text {
				c.ws.WriteMessage(websocket.TextMessage, msg.Bytes())
			} else {
				c.ws.WriteMessage(websocket.BinaryMessage, msg.Bytes())
			}
		} else {
			panic("Output channel of connection broken.")
		}
	}
}

func (c *connection) close() {
	if c.user != "" {
		c.server.Unregister(c, c.user+"@")
	} else if c.remote_domain != "" {
		c.server.Unregister(c, "@"+c.remote_domain)
	}
	c.ws.Close()
}

func (c *connection) send(msg Message) {
	c.out <- msg
}

func (c *connection) SendRequest(rt RequestType, url *Url, headers map[string]string, body []byte) (*Response, error) {
	seq := atomic.AddUint64(&c.currentSeq, uint64(1))
	req := NewRequest(rt, url, int(seq), headers, body)

	c.pendingRequestsLock.Lock()
	c.pendingRequests[seq] = make(chan *Response)
	c.pendingRequestsLock.Unlock()
	log.Println("Sending request")
	c.send(req)
	var (
		resp    *Response
		ok      bool = false
		timeout bool = false
	)
	select {
	case resp, ok = <-c.pendingRequests[seq]:
	case <-time.After(time.Second * 15):
		timeout = true
	}
	log.Println("Received response or timeout")

	c.pendingRequestsLock.Lock()
	delete(c.pendingRequests, seq)
	c.pendingRequestsLock.Unlock()

	if !ok {
		log.Println("Something went wrong when reading channel")
		return nil, errors.New("Error when receiving response")
	}
	if timeout {
		log.Println("Request timed out")
		return nil, errors.New("Request timed out")
	}
	log.Println("Return response")
	return resp, nil
}

func (c *connection) checkState(msg Message) {
	// If this connection is negotiated and authenticated we normaly handle the message
	if c.negotiated && c.authenticated {
		c.handleMessage(msg)
	} else if req, ok := msg.(*Request); ok {
		c.bootstrap(req)
	} else {
		//Invalid state
	}
}
