package main

import (
  "github.com/gorilla/websocket"
  "log"
  "encoding/json"
  "errors"
)

type connection struct {
  ws *websocket.Conn
  database *Database
  server *server

  negotiated bool
  authenticated bool
  user string
  remote_domain string
}

func NewConnection(ws *websocket.Conn, db *Database, srv *server) *connection {
  if ws == nil || db == nil || srv == nil {
    panic("Cannot initialize fosp connection without websocket, database or server")
  }
  con := new(connection)
  con.ws = ws
  con.database = db
  con.server = srv
  return con
}

func (c *connection) listen() {
  for {
    _, message, err := c.ws.ReadMessage()
    if err != nil {
      log.Println("Error while receiving new WebSocket message :: ", err)
      c.ws.Close()
      break
    }
    if msg, err := parseMessage(string(message)); err != nil {
      log.Println("Error while parsing message :: ", err)
      //c.ws.Close()
      //break
    } else {
      c.handleMessage(msg)
    }
  }
}

func (c *connection) handleMessage(msg Message) {
  req, is_request := msg.(*Request)
  if is_request && ! c.negotiated {
    if err := c.negotiate(req); err != nil {
      //c.ws.Close()
      //break
    }
  } else if is_request && req.request == Register {
    if err := c.register(req); err != nil {
      //c.ws.Close()
      //break
    }
  } else if is_request && ! c.authenticated {
    if err := c.authenticate(req); err != nil {
      //c.ws.Close()
      //break
    } else {
      c.server.registerConnection(c, c.user)
    }
  } else {
    if c.user != "" {
      if is_request {
        resp := c.handleRequest(req)
        c.send(resp)
      }
    } else if c.remote_domain != "" {

    }
  }
}

func (c *connection) send(msg Message) {
  c.ws.WriteMessage(websocket.TextMessage, msg.Bytes())
}

func (c *connection) negotiate(req *Request) error {
  if req.request != Connect {
    return errors.New("Recieved message on not negotiated connection")
  }
  var obj ConnectionNegotiationObject
  err := json.Unmarshal([]byte(req.body), &obj)
  if err != nil {
    return err
  } else if obj.Version != "0.1" {
    c.send(req.Failed(400, "Version not supported"))
    return errors.New("Unsupported FOSP version :: " + obj.Version)
  } else {
    c.negotiated = true
    c.send(req.Succeeded(200, ""))
    return nil
  }
}

func (c *connection) authenticate(req *Request) error {
  if req.request != Authenticate {
    return errors.New("Recieved message on not authenticated connection")
  }
  var obj AuthenticationObject
  err := json.Unmarshal([]byte(req.body), &obj)
  if err != nil {
    return err
  } else if obj.Name == "" || obj.Password == "" {
    c.send(req.Failed(400, "Name or password missing"))
    return errors.New("Name of password missing")
  } else {
    if err := c.database.Authenticate(obj.Name, obj.Password); err == nil {
      c.authenticated = true
      c.user = obj.Name + "@" + c.server.GetDomain()
      c.send(req.Succeeded(200, ""))
      return nil
    } else {
      c.send(req.Failed(403, "Invalid user or password"))
      return nil
    }
  }
}

func (c *connection) register(req *Request) error {
  if req.request != Register {
    log.Fatal("Tried to register but request is not a REGISTER request")
  }
  var obj AuthenticationObject
  err := json.Unmarshal([]byte(req.body), &obj)
  if err != nil {
    return err
  } else if obj.Name == "" || obj.Password == "" {
    c.send(req.Failed(400, "Name or password missing"))
    return errors.New("Name of password missing")
  } else {
    if err := c.database.Register(obj.Name, obj.Password); err == nil {
      c.send(req.Succeeded(200, ""))
      return nil
    } else {
      c.send(req.Failed(500, err.Error()))
      return nil
    }
  }
}
