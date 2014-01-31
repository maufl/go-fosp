package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
)

type server struct {
	database    *database
	connections map[string][]*connection
	connsLock   sync.Mutex
	domain      string
}

func NewServer(dbDriver DatabaseDriver, domain string) *server {
	if dbDriver == nil {
		panic("Cannot initialize server without database")
	}
	s := new(server)
	s.database = NewDatabase(dbDriver, s)
	s.domain = domain
	s.connections = make(map[string][]*connection)
	return s
}

func (s *server) requestHandler(res http.ResponseWriter, req *http.Request) {
	ws, err := websocket.Upgrade(res, req, nil, 1024, 104)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(res, "Not a WebSocket handshake", 400)
		return
	} else if err != nil {
		log.Println("Error while setting up WebSocket connection :: ", err)
		return
	}
	log.Println("Successfully accepted new connection")
	NewConnection(ws, s)
}

func (s *server) registerConnection(c *connection, remote string) {
	s.connsLock.Lock()
	s.connections[remote] = append(s.connections[remote], c)
	s.connsLock.Unlock()
}

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

func (s *server) routeNotification(user string, notf *Notification) {
	if strings.HasSuffix(user, "@"+s.domain) {
		for _, connection := range s.connections[user] {
			connection.send(notf)
		}
	}
}

func (s *server) forwardRequest(user string, rt RequestType, url *Url, headers map[string]string, body string) (*Response, error) {
	remote_domain := url.Domain()
	headers["User"] = user
	var remote_connection *connection
	if connections, ok := s.connections["@"+remote_domain]; ok {
		for _, connection := range connections {
			if connection != nil {
				remote_connection = connection
			}
		}
	} else if remote_connection == nil {
		var err error
		remote_connection, err = OpenConnection(s, remote_domain)
		if err != nil {
			return nil, err
		}
	}
	resp, err := remote_connection.SendRequest(rt, url, headers, body)
	log.Println("Recieved response from forwarded request")
	if err != nil {
		log.Println("Error occured while forwarding " + err.Error())
		return nil, err
	} else {
		resp.DeleteHead("User")
		return resp, nil
	}
}

func (s *server) Domain() string {
	if s.domain == "" {
		return "localhost.localdomain"
	} else {
		return s.domain
	}
}
