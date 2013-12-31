package main

import (
  "github.com/gorilla/websocket"
  "net/http"
  "log"
  "strings"
)

type server struct {
  driver DatabaseDriver
  database *Database
  connections map[string][]*connection
  domain string
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
  c := &connection{ ws: ws, database: s.database, server: s, negotiated: false, authenticated: false }
  c.listen()
}

func (s *server) registerConnection(c *connection, remote string) {
  s.connections[remote] = append(s.connections[remote], c)
}

func (s *server) routeMessage(user string, m Message) {
  if strings.HasSuffix(user, "@" + s.domain) {
    for _, connection := range s.connections[user] {
      connection.send(m)
    }
  }
}

func (s *server) GetDomain() string {
  if s.domain == "" {
    return "localhost.localdomain"
  } else {
    return s.domain
  }
}
