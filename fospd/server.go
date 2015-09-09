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
	"github.com/maufl/go-fosp/fosp"
	"github.com/op/go-logging"
	"net/http"
	"strings"
	"sync"
)

var srvLog = logging.MustGetLogger("go-fosp/fosp/server")

// Server represents a FOSP server.
// It is responsible for a single domain, uses a database to store the data
// and manages the Connections.
type Server struct {
	database        *Database
	connections     map[string][]*ServerConnection
	connectionsLock sync.RWMutex
	domain          string
}

// NewServer initializes a new server struct and returns it.
// The DatabaseDriver will be used by this Server to store data.
// The Server will process requests for resources on the provided domain.
// The RequestHandler method of the Server can be used as a handler for incoming WebSocket connections
func NewServer(dbDriver DatabaseDriver, domain string) *Server {
	if dbDriver == nil {
		panic("Cannot initialize server without database")
	}
	s := new(Server)
	s.database = NewDatabase(dbDriver, s)
	s.domain = domain
	s.connections = make(map[string][]*ServerConnection)
	return s
}

// RequestHandler is a method that accepts a HTTP request and tries to upgrade it to a WebSocket connection.
// On success it instanciates a new fosp.Connection using the new WebSocket connection.
func (s *Server) RequestHandler(res http.ResponseWriter, req *http.Request) {
	srvLog.Debug("Recieved a new http request %s, %s\n%s", req.Method, req.URL.String(), req.Header)
	ws, err := websocket.Upgrade(res, req, nil, 1024, 104)
	if _, ok := err.(websocket.HandshakeError); ok {
		srvLog.Warning("Recieved a request that is not a Websocket handshake")
		http.Error(res, "Not a WebSocket handshake", 400)
		return
	} else if err != nil {
		srvLog.Warning("Error while setting up WebSocket connection: %s", err)
		return
	}
	srvLog.Notice("Successfully accepted new connection")
	NewServerConnection(ws, s)
}

// registerConnection registers a connection with the Server for a remote entity.
// The Server saves this connection to it's mapping and associates it with the given remote entity.
func (s *Server) registerConnection(c *ServerConnection, remote string) {
	s.connectionsLock.Lock()
	s.connections[remote] = append(s.connections[remote], c)
	s.connectionsLock.Unlock()
}

// Unregister removes an connection from the list of known connections of this Server.
// When no such connection is known by the Server then this is a nop.
func (s *Server) Unregister(c *ServerConnection, remote string) {
	s.connectionsLock.Lock()
	for i, v := range s.connections[remote] {
		if v == c {
			s.connections[remote] = append(s.connections[remote][:i], s.connections[remote][i+1:]...)
			break
		}
	}
	s.connectionsLock.Unlock()
}

// routeNotification routes a notification to a user.
// It first determins if the user belongs to the domain of the Server.
// If that's the case, it searches it's known connection for a connection to this user and sends the notification.
// Else it routes the notification to a remote server, opening a new connection if necessary.
func (s *Server) routeNotification(user string, notf *fosp.Notification) {
	srvLog.Info("Sending notification %v to user %s", notf, user)
	if strings.HasSuffix(user, "@"+s.domain) {
		userName := strings.TrimSuffix(user, s.domain)
		srvLog.Debug("Is local user %s", userName)
		s.connectionsLock.RLock()
		srvLog.Debug("Connections are %v", s.connections[userName])
		for _, connection := range s.connections[userName] {
			srvLog.Debug("Sending notification on local connection")
			connection.Send(notf)
		}
		s.connectionsLock.RUnlock()
	} else if notf.URL().Domain() == s.domain {
		parts := strings.Split(user, "@")
		if len(parts) != 2 {
			panic(user + " is not a valid user identifier")
		}
		remoteDomain := parts[1]
		srvLog.Debug("Is local notification that will be routed to remote server")
		remoteConnection, err := s.getOrOpenRemoteConnection(remoteDomain)
		if err == nil {
			notf.SetHead("User", user)
			remoteConnection.Send(notf)
		}
	}
}

// forwardRequest sends a request to a remote Server and returns the response or an error.
// It is used to forward a request from a local user for a non local resources to remote servers.
func (s *Server) forwardRequest(user string, rt fosp.RequestType, url *fosp.URL, headers map[string]string, body []byte) (*fosp.Response, error) {
	remoteDomain := url.Domain()
	headers["User"] = user
	remoteConnection, err := s.getOrOpenRemoteConnection(remoteDomain)
	if err != nil {
		return nil, err
	}
	resp, err := remoteConnection.SendRequest(rt, url, headers, body)
	srvLog.Info("Recieved response from forwarded request")
	if err != nil {
		srvLog.Warning("Error occured while forwarding " + err.Error())
		return nil, err
	}
	resp.DeleteHead("User")
	return resp, nil
}

// getOrOpenRemoteConnection returns a connection to the remoteDomain.
// If such a connection already exists and is known to the Server, it is reused.
// Otherwise a new connection is opened.
// If a new connection is opened, the call will be blocked until the new connection is authenticated or failed.
func (s *Server) getOrOpenRemoteConnection(remoteDomain string) (*ServerConnection, error) {
	s.connectionsLock.RLock()
	if connections, ok := s.connections["@"+remoteDomain]; ok {
		for _, connection := range connections {
			s.connectionsLock.RUnlock()
			return connection, nil
		}
	}
	s.connectionsLock.RUnlock()
	return OpenServerConnection(s, remoteDomain)
}

// Domain returns the domain this Server handles.
func (s *Server) Domain() string {
	if s.domain == "" {
		return "localhost.localdomain"
	}
	return s.domain
}
