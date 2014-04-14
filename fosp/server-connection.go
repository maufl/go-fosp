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
	"net/http"
)

type ServerConnection struct {
	Connection
	server *server

	user          string
	remote_domain string
}

func NewServerConnection(ws *websocket.Conn, srv *server) *ServerConnection {
	if ws == nil || srv == nil {
		panic("Cannot initialize fosp connection without websocket or server")
	}
	con := &ServerConnection{*NewConnection(ws), srv, "", ""}
	con.lg = logging.MustGetLogger("go-fosp/fosp/server-connection")
	con.messageHandler = con.checkState
	go con.listen()
	go con.talk()
	return con
}

func OpenServerConnection(srv *server, remote_domain string) (*ServerConnection, error) {
	url := "ws://" + remote_domain + ":1337"
	srv.lg.Info("Opening new connection to %s", url)
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		srv.lg.Error("Error when opening new WebSocket connection %s", err)
		return nil, err
	}
	connection := NewServerConnection(ws, srv)
	connection.negotiated = true
	connection.authenticated = true
	connection.remote_domain = remote_domain
	resp, err := connection.SendRequest(Connect, &Url{}, map[string]string{}, []byte("{\"version\":\"0.1\"}"))
	if err != nil {
		return nil, errors.New("Error when negotiating connection")
	} else if resp.response != Succeeded {
		connection.lg.Warning("Connection negotiation failed!")
		return nil, errors.New("Connection negotiation failed!")
	}
	connection.lg.Info("Connection successfully negotiated")
	resp, err = connection.SendRequest(Authenticate, &Url{}, map[string]string{}, []byte("{\"type\":\"server\", \"domain\":\""+srv.Domain()+"\"}"))
	if err != nil || resp.response != Succeeded {
		connection.lg.Warning("Error when authenticating")
		return nil, errors.New("Error when authenticating")
	}
	connection.lg.Info("Successfully authenticated")
	srv.registerConnection(connection, "@"+remote_domain)
	return connection, nil
}

// Close this connection and clean up
// TODO: Websocket should send close message before tearing down the connection
func (c *ServerConnection) Close() {
	if c.user != "" {
		c.server.Unregister(c, c.user+"@")
	} else if c.remote_domain != "" {
		c.server.Unregister(c, "@"+c.remote_domain)
	}
	c.ws.Close()
}

func (c *ServerConnection) checkState(msg Message) {
	// If this connection is negotiated and authenticated we normaly handle the message
	if c.negotiated && c.authenticated {
		c.handleMessage(msg)
	} else if req, ok := msg.(*Request); ok {
		c.bootstrap(req)
	} else {
		// TODO: Invalid state
	}
}
