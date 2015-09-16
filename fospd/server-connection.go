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
	"github.com/gorilla/websocket"
	"github.com/maufl/go-fosp/fosp/fospws"
	"github.com/op/go-logging"
	"net/http"
)

var servConnLog = logging.MustGetLogger("go-fosp/fosp/server-connection")

// ServerConnection represents a FOSP connection in the server.
type ServerConnection struct {
	*fospws.Connection
	server *Server

	SaslMechanism string

	User         string
	RemoteDomain string
}

// NewServerConnection creates a new ServerConnection for an existing WebSocket connection.
func NewServerConnection(ws *websocket.Conn, srv *Server) *ServerConnection {
	if ws == nil || srv == nil {
		panic("Cannot initialize fosp connection without websocket or server")
	}
	con := &ServerConnection{Connection: fospws.NewConnection(ws), server: srv, User: "", RemoteDomain: ""}
	con.RegisterMessageHandler(con)
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
	connection.RemoteDomain = remoteDomain
	srv.registerConnection(connection, "@"+remoteDomain)
	return connection, nil
}

// Close this connection and clean up
// TODO: Websocket should send close message before tearing down the connection
func (c *ServerConnection) Close() {
	if c.User != "" {
		c.server.Unregister(c, c.User+"@")
	} else if c.RemoteDomain != "" {
		c.server.Unregister(c, "@"+c.RemoteDomain)
	}
	c.Connection.Close()
}
