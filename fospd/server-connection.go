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

package main

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/maufl/go-fosp/fosp"
	"github.com/op/go-logging"
	"net/http"
	"sync/atomic"
)

// Constants which denote the state of a connection
const (
	Opened uint32 = iota
	Negotiated
	Authenticated
	Closing
	Closed
)

// ErrNegotiationFailed is returned when the negotiation of a new connection failed.
var ErrNegotiationFailed = errors.New("negotiation failed")

var servConnLog = logging.MustGetLogger("go-fosp/fosp/server-connection")

// ServerConnection represents a FOSP connection in the server.
type ServerConnection struct {
	fosp.Connection
	server *Server

	state uint32

	user         string
	remoteDomain string
}

// NewServerConnection creates a new ServerConnection for an existing WebSocket connection.
func NewServerConnection(ws *websocket.Conn, srv *Server) *ServerConnection {
	if ws == nil || srv == nil {
		panic("Cannot initialize fosp connection without websocket or server")
	}
	con := &ServerConnection{Connection: *fosp.NewConnection(ws), server: srv, user: "", remoteDomain: ""}
	con.RegisterMessageHandler(con)
	go con.Listen()
	go con.Talk()
	return con
}

// OpenServerConnection opens a new ServerConnection to the remoteDomain.
// It first opens a WebSocket connection to the remoteDomain.
// Then it negotiates the connection parameters and authenticates.
// If any of the steps fail, nil and an error is returned.
func OpenServerConnection(srv *Server, remoteDomain string) (*ServerConnection, error) {
	url := "ws://" + remoteDomain + ":1337"
	srvLog.Info("Opening new connection to %s", url)
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		srvLog.Error("Error when opening new WebSocket connection " + err.Error())
		return nil, err
	}
	connection := NewServerConnection(ws, srv)
	connection.state = Authenticated
	connection.remoteDomain = remoteDomain
	resp, err := connection.SendRequest(fosp.Connect, &url.URL{}, map[string]string{}, []byte("{\"version\":\"0.1\"}"))
	if err != nil {
		return nil, err
	} else if resp.ResponseType() != fosp.Succeeded {
		servConnLog.Warning("Connection negotiation failed!")
		return nil, ErrNegotiationFailed
	}
	servConnLog.Info("Connection successfully negotiated")
	resp, err = connection.SendRequest(fosp.Authenticate, &url.URL{}, map[string]string{}, []byte("{\"type\":\"server\", \"domain\":\""+srv.Domain()+"\"}"))
	if err != nil || resp.ResponseType() != fosp.Succeeded {
		servConnLog.Warning("Error when authenticating")
		return nil, fosp.ErrAuthenticationFailed
	}
	servConnLog.Info("Successfully authenticated")
	srv.registerConnection(connection, "@"+remoteDomain)
	return connection, nil
}

// Close this connection and clean up
// TODO: Websocket should send close message before tearing down the connection
func (c *ServerConnection) Close() {
	if c.user != "" {
		c.server.Unregister(c, c.user+"@")
	} else if c.remoteDomain != "" {
		c.server.Unregister(c, "@"+c.remoteDomain)
	}
	c.Ws.Close()
}

// HandleMessage is the entrypoint for processing all messages.
func (c *ServerConnection) HandleMessage(msg fosp.Message) {
	// If this connection is negotiated and authenticated we normaly handle the message
	if atomic.CompareAndSwapUint32(&c.state, Authenticated, Authenticated) {
		c.handleMessage(msg)
	} else if req, ok := msg.(*fosp.Request); ok {
		c.bootstrap(req)
	} else {
		// TODO: Invalid state
	}
}
