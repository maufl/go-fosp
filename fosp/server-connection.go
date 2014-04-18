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
)

// ServerConnection represents a FOSP connection in the server.
type ServerConnection struct {
	Connection
	server *Server

	user         string
	remoteDomain string
}

// NewServerConnection creates a new ServerConnection for an existing WebSocket connection.
func NewServerConnection(ws *websocket.Conn, srv *Server) *ServerConnection {
	if ws == nil || srv == nil {
		panic("Cannot initialize fosp connection without websocket or server")
	}
	con := &ServerConnection{Connection{ws: ws, pendingRequests: make(map[uint64]chan *Response), out: make(chan Message)}, srv, "", ""}
	con.lg = logging.MustGetLogger("go-fosp/fosp/server-connection")
	con.RegisterMessageHandler(con)
	go con.listen()
	go con.talk()
	return con
}

// OpenServerConnection opens a new ServerConnection to the remoteDomain.
// It first opens a WebSocket connection to the remoteDomain.
// Then it negotiates the connection parameters and authenticates.
// If any of the steps fail, nil and an error is returned.
func OpenServerConnection(srv *Server, remoteDomain string) (*ServerConnection, error) {
	url := "ws://" + remoteDomain + ":1337"
	srv.lg.Info("Opening new connection to %s", url)
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		srv.lg.Error("Error when opening new WebSocket connection %s", err)
		return nil, err
	}
	connection := NewServerConnection(ws, srv)
	connection.negotiated = true
	connection.authenticated = true
	connection.remoteDomain = remoteDomain
	resp, err := connection.SendRequest(Connect, &URL{}, map[string]string{}, []byte("{\"version\":\"0.1\"}"))
	if err != nil {
		return nil, errors.New("error when negotiating connection")
	} else if resp.response != Succeeded {
		connection.lg.Warning("Connection negotiation failed!")
		return nil, errors.New("connection negotiation failed")
	}
	connection.lg.Info("Connection successfully negotiated")
	resp, err = connection.SendRequest(Authenticate, &URL{}, map[string]string{}, []byte("{\"type\":\"server\", \"domain\":\""+srv.Domain()+"\"}"))
	if err != nil || resp.response != Succeeded {
		connection.lg.Warning("Error when authenticating")
		return nil, errors.New("error when authenticating")
	}
	connection.lg.Info("Successfully authenticated")
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
	c.ws.Close()
}

// HandleMessage is the entrypoint for processing all messages.
func (c *ServerConnection) HandleMessage(msg Message) {
	// If this connection is negotiated and authenticated we normaly handle the message
	if c.negotiated && c.authenticated {
		c.handleMessage(msg)
	} else if req, ok := msg.(*Request); ok {
		c.bootstrap(req)
	} else {
		// TODO: Invalid state
	}
}
