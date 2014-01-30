package main

import (
  "github.com/gorilla/websocket"
  "log"
  _ "encoding/json"
  "errors"
  "sync/atomic"
  "sync"
  "time"
  "net/http"
)

type connection struct {
  ws *websocket.Conn
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
  go con.listen()
  go con.talk()
  return con
}

func OpenConnection(srv *server, remote_domain string) (*connection, error) {
  url := "ws://"+remote_domain+":1337"
  log.Println("Opening new connection to " + url)
  ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
  if err != nil {
    return nil, err
  }
  connection := NewConnection(ws, srv)
  resp, err := connection.SendRequest(Connect, &Url{}, map[string]string{}, "{\"version\":0.1}")
  if err != nil || resp.response != Succeeded {
    return nil, errors.New("Error when negotiating connection")
  }
  log.Println("Successfully negotiated")
  connection.negotiated = true
  resp, err = connection.SendRequest(Authenticate, &Url{}, map[string]string{}, "{\"type\":\"server\", \"domain\":\"" + srv.Domain() + "\"}")
  if err != nil || resp.response != Succeeded {
    return nil, errors.New("Error when authenticating")
  }
  log.Println("Successfully authenticated")
  connection.authenticated = true
  connection.remote_domain = remote_domain
  srv.registerConnection(connection, "@" + remote_domain)
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
    if msg, err := parseMessage(string(message)); err != nil {
      log.Println("Error while parsing message :: ", err)
      c.close()
      break
    } else {
      log.Println("Received new message")
      c.checkState(msg)
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
  if c.user != "" {
    c.server.Unregister(c, c.user + "@")
  } else if c.remote_domain != "" {
    c.server.Unregister(c, "@" + c.remote_domain)
  }
  c.ws.Close()
}


func (c *connection) send(msg Message) {
  c.out <- msg
}

func (c *connection) SendRequest(rt RequestType, url *Url, headers map[string]string, body string) (*Response, error) {
  seq := atomic.AddUint64(&c.currentSeq, uint64(1))
  req := NewRequest(rt, url, int(seq), headers, body)

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
