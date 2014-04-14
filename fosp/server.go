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
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"net/http"
	"strings"
	"sync"
)

var lg = logging.MustGetLogger("go-fosp/fosp")

type server struct {
	database    *database
	connections map[string][]*connection
	// BUG(maufl): Should use RWLock and also lock on reading
	connsLock sync.Mutex
	domain    string
	lg        *logging.Logger
}

// NewServer initializes a new server struct and returns it.
// The DatabaseDriver will be used by this server to store data.
// The server will process requests for resources on the provided domain.
// The RequestHandler method of the server can be used as a handler for incoming WebSocket connections
func NewServer(dbDriver DatabaseDriver, domain string) *server {
	if dbDriver == nil {
		panic("Cannot initialize server without database")
	}
	s := new(server)
	s.database = NewDatabase(dbDriver, s)
	s.domain = domain
	s.connections = make(map[string][]*connection)
	s.lg = logging.MustGetLogger("go-fosp/fosp/server")
	return s
}

// The RequestHandler method accepts a HTTP request and tries to upgrade it to a WebSocket connection.
// On success it instanciates a new fosp.Connection using the new WebSocket connection.
func (s *server) RequestHandler(res http.ResponseWriter, req *http.Request) {
	ws, err := websocket.Upgrade(res, req, nil, 1024, 104)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(res, "Not a WebSocket handshake", 400)
		return
	} else if err != nil {
		s.lg.Warning("Error while setting up WebSocket connection: %s", err)
		return
	}
	s.lg.Notice("Successfully accepted new connection")
	NewConnection(ws, s)
}

// registerConnection registers a connection with the server for a remote entity.
// The server saves this connection to it's mapping and associates it with the given remote entity.
func (s *server) registerConnection(c *connection, remote string) {
	s.connsLock.Lock()
	s.connections[remote] = append(s.connections[remote], c)
	s.connsLock.Unlock()
}

// Unregister removes an connection from the list of known connections of this server.
// When no such connection is known by the server then this is a nop.
func (s *server) Unregister(c *connection, remote string) {
	s.connsLock.Lock()
	for i, v := range s.connections[remote] {
		if v == c {
			s.connections[remote] = append(s.connections[remote][:i], s.connections[remote][i+1:]...)
			break
		}
	}
	s.connsLock.Unlock()
}

// routeNotification routes a notification to a user.
// It first determins if the user belongs to the domain of the server.
// If that's the case, it searches it's known connection for a connection to this user and sends the notification.
// Else it routes the notification to a remote server, opening a new connection if necessary.
func (s *server) routeNotification(user string, notf *Notification) {
	s.lg.Info("Sending notification %v to user %s", notf, user)
	if strings.HasSuffix(user, "@"+s.domain) {
		user_name := strings.TrimSuffix(user, s.domain)
		s.lg.Debug("Is local user %s", user_name)
		s.lg.Debug("Connections are %v", s.connections[user_name])
		for _, connection := range s.connections[user_name] {
			s.lg.Debug("Sending notification on local connection")
			connection.send(notf)
		}
	} else if notf.url.Domain() == s.domain {
		parts := strings.Split(user, "@")
		if len(parts) != 2 {
			panic(user + " is not a valid user identifier")
		}
		remote_domain := parts[1]
		s.lg.Debug("Is local notification that will be routed to remote server")
		remote_connection, err := s.getOrOpenRemoteConnection(remote_domain)
		if err == nil {
			notf.SetHead("User", user)
			remote_connection.send(notf)
		}
	}
}

// forwardRequest sends a request to a remote server and returns the response or an error.
// It is used to forward a request from a local user for a non local resources to remote servers.
func (s *server) forwardRequest(user string, rt RequestType, url *Url, headers map[string]string, body []byte) (*Response, error) {
	remote_domain := url.Domain()
	headers["User"] = user
	remote_connection, err := s.getOrOpenRemoteConnection(remote_domain)
	if err != nil {
		return nil, err
	}
	resp, err := remote_connection.SendRequest(rt, url, headers, body)
	s.lg.Info("Recieved response from forwarded request")
	if err != nil {
		s.lg.Warning("Error occured while forwarding " + err.Error())
		return nil, err
	} else {
		resp.DeleteHead("User")
		return resp, nil
	}
}

// getOrOpenRemoteConnection returns a connection to the remote_domain.
// If such a connection already exists and is known to the server, it is reused.
// Otherwise a new connection is opened.
// If a new connection is opened, the call will be blocked until the new connection is authenticated or failed.
func (s *server) getOrOpenRemoteConnection(remote_domain string) (*connection, error) {
	if connections, ok := s.connections["@"+remote_domain]; ok {
		for _, connection := range connections {
			return connection, nil
		}
	}
	return OpenConnection(s, remote_domain)
}

// Domain returns the domain this server handles.
func (s *server) Domain() string {
	if s.domain == "" {
		return "localhost.localdomain"
	} else {
		return s.domain
	}
}
