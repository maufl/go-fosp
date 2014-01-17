package main

import (
  _ "github.com/gorilla/websocket"
  "log"
  "encoding/json"
  "errors"
  _ "sync/atomic"
)

func (c *connection) handleMessage(msg Message) {
  // If this connection is negotiated and authenticated we normaly handle the message
  if c.negotiated && c.authenticated {
    c.switchMessage(msg)
  } else if req, ok := msg.(*Request); ok {
    c.bootstrap(req)
  } else {
    //Invalid state
  }
}

func (c *connection) switchMessage(msg Message) {
  if req, ok := msg.(*Request); ok {
    resp := c.handleRequest(req)
    c.send(resp)
  }
  if resp, ok := msg.(*Response); ok {
    c.handleResponse(resp)
  }
  if ntf, ok := msg.(*Notification); ok {
    c.handleNotification(ntf)
  }
}

func (c *connection) bootstrap(req *Request) {
  if ! c.negotiated {
    if err := c.negotiate(req); err != nil {
      //c.ws.Close()
      //break
    }
  } else if req.request == Register {
    if err := c.register(req); err != nil {
      //c.ws.Close()
      //break
    }
  } else if ! c.authenticated {
    if err := c.authenticate(req); err != nil {
      //c.ws.Close()
      //break
    } else {
      c.server.registerConnection(c, c.user)
    }
  } else {
    //Invalid state
  }
}

func (c *connection) handleResponse(resp *Response) {
  if ch, ok := c.pendingRequests[uint64(resp.seq)]; ok {
    ch <- resp
  }
}

func (c *connection) handleNotification(ntf *Notification) {

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
