package main

import (
  "github.com/gorilla/websocket"
  "log"
  _ "encoding/json"
  "errors"
  "sync/atomic"
  "sync"
  "time"
)

type connection struct {
  ws *websocket.Conn
  database *database
  server *server

  negotiated bool
  authenticated bool
  user string
  remote_domain string

  currentSeq uint64
  pendingRequests map[uint64]chan *Response
  pendingRequestsLock sync.Mutex

  out chan Message
}

func NewConnection(ws *websocket.Conn, db *database, srv *server) *connection {
  if ws == nil || db == nil || srv == nil {
    panic("Cannot initialize fosp connection without websocket, database or server")
  }
  con := new(connection)
  con.ws = ws
  con.database = db
  con.server = srv
  con.currentSeq = 0
  con.pendingRequests = make(map[uint64]chan *Response)
  con.out = make(chan Message)
  go con.listen()
  go con.talk()
  return con
}

func (c *connection) listen() {
  for {
    _, message, err := c.ws.ReadMessage()
    if err != nil {
      log.Println("Error while receiving new WebSocket message :: ", err)
      c.close()
      break
    }
    if msg, err := parseMessage(string(message)); err != nil {
      log.Println("Error while parsing message :: ", err)
      c.close()
      break
    } else {
      c.handleMessage(msg)
    }
  }
}

func (c *connection) talk() {
  for {
    if msg, ok := <-c.out; ok {
      c.ws.WriteMessage(websocket.TextMessage, msg.Bytes())
    } else {
      println("Output channel of connection broken.")
      break
    }
  }
}

func (c *connection) close() {
  c.server.Unregister(c)
  c.ws.Close()
}


func (c *connection) send(msg Message) {
  c.out <- msg
}

func (c *connection) SendRequest(rt RequestType, url Url, headers map[string]string, body string) (*Response, error) {
  seq := atomic.AddUint64(&c.currentSeq, uint64(1))
  req := NewRequest(rt, &url, int(seq), headers, body)

  c.pendingRequestsLock.Lock()
  c.pendingRequests[seq] = make(chan *Response)
  c.pendingRequestsLock.Unlock()

  c.send(req)
  var (
    resp *Response
    ok bool = false
    timeout bool = false
  )
  select {
  case resp, ok = <-c.pendingRequests[seq]:
  case <-time.After(time.Second * 15):
    timeout = true
  }

  c.pendingRequestsLock.Lock()
  delete(c.pendingRequests, seq)
  c.pendingRequestsLock.Unlock()

  if !ok {
    return nil, errors.New("Error when receiving response")
  }
  if timeout {
    return nil, errors.New("Request timed out")
  }
  return resp, nil
}
